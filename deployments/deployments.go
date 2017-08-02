package deployments

import (
	"fmt"
	"github.com/ricjcosme/kube-monkey/config"
	kube "k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.5/pkg/labels"
	"k8s.io/client-go/1.5/pkg/selection"
	"k8s.io/client-go/1.5/pkg/util/sets"
	"strconv"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"time"
	"strings"
	"log"
)

type Deployment struct {
	name       string
	namespace  string
	identifier string
	mtbf       int
}

// Create a new instance of Deployment
func New(dep *v1beta1.Deployment) (*Deployment, error) {
	ident, err := identifier(dep)
	if err != nil {
		return nil, err
	}
	mtbf, err := meanTimeBetweenFailures(dep)
	if err != nil {
		return nil, err
	}

	return &Deployment{
		name:       dep.Name,
		namespace:  dep.Namespace,
		identifier: ident,
		mtbf:       mtbf,
	}, nil
}

// Returns the value of the label defined by config.IdentLabelKey
// from the deployment labels
// This label should be unique to a deployment, and is used to
// identify the pods that belong to this deployment, as pods
// inherit labels from the Deployment
func identifier(kubedep *v1beta1.Deployment) (string, error) {
	identifier, ok := kubedep.Labels[config.IdentLabelKey]
	if !ok {
		return "", fmt.Errorf("Deployment %s does not have %s label", kubedep.Name, config.IdentLabelKey)
	}
	return identifier, nil
}

// Read the mean-time-between-failures value defined by the Deployment
// in the label defined by config.MtbfLabelKey
func meanTimeBetweenFailures(kubedep *v1beta1.Deployment) (int, error) {
	mtbf, ok := kubedep.Labels[config.MtbfLabelKey]
	if !ok {
		return -1, fmt.Errorf("Deployment %s does not have %s label", kubedep.Name, config.MtbfLabelKey)
	}

	mtbfInt, err := strconv.Atoi(mtbf)
	if err != nil {
		return -1, err
	}

	if !(mtbfInt > 0) {
		return -1, fmt.Errorf("Invalid value for label %s: %d", config.MtbfLabelKey, mtbfInt)
	}

	return mtbfInt, nil
}

func (d *Deployment) Name() string {
	return d.name
}

func (d *Deployment) Namespace() string {
	return d.namespace
}

func (d *Deployment) Mtbf() int {
	return d.mtbf
}

// Returns a list of running pods for the deployment
func (d *Deployment) RunningPods(client *kube.Clientset) ([]v1.Pod, error) {
	runningPods := []v1.Pod{}

	pods, err := d.Pods(client)
	if err != nil {
		return nil, err
	}

	for _, pod := range pods {
		if pod.Status.Phase == v1.PodRunning {
			runningPods = append(runningPods, pod)
		}
	}

	return runningPods, nil
}

// Returns a list of pods under the Deployment
func (d *Deployment) Pods(client *kube.Clientset) ([]v1.Pod, error) {
	labelSelector, err := d.LabelFilterForPods()
	if err != nil {
		return nil, err
	}

	podlist, err := client.Core().Pods(d.namespace).List(*labelSelector)
	if err != nil {
		return nil, err
	}
	return podlist.Items, nil
}

func (d *Deployment) DeletePod(client *kube.Clientset, pod v1.Pod) error {
	deleteopts := &api.DeleteOptions{
		GracePeriodSeconds: config.GracePeriodSeconds(),
	}

	// K8s event definition
	var message []string = []string{config.KubeMonkeyAppName(), "killed pod", pod.Name, "in deployment", d.name}
	ev := &v1.Event{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Event",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: strings.Join([]string{config.KubeMonkeyAppName(), "-", pod.Name},""),
		},
		InvolvedObject: v1.ObjectReference{
			Kind: "Pod",
			Namespace: d.namespace,
			Name: pod.Name,
			UID: pod.GetUID(),
			APIVersion: pod.APIVersion,
			ResourceVersion: pod.ResourceVersion,
		},
		Type: "Warning",
		Reason: "Chaos",
		Message: strings.Join(message, " "),
		Source: v1.EventSource{config.KubeMonkeyAppName(), pod.ClusterName},
		FirstTimestamp: unversioned.Time{time.Now()},
		LastTimestamp: unversioned.Time{time.Now()},
		Count: 1,
	}

	error := client.Core().Pods(d.namespace).Delete(pod.Name, deleteopts)

	if error != nil {
		// Oops, something went wrong with pod deletion
		log.Fatal(error)
	} else {
		// K8s event creation - pod was killed
		_, err := client.Core().Events(d.namespace).Create(ev)
		if err != nil {
			fmt.Println(err)
		}
	}

	return error
}

// Create a label filter to filter only for pods that belong to the this
// deployment. This is done using the identifier label
func (d *Deployment) LabelFilterForPods() (*api.ListOptions, error) {
	req, err := d.LabelRequirementForPods()
	if err != nil {
		return nil, err
	}
	labelFilter := &api.ListOptions{
		LabelSelector: labels.NewSelector().Add(*req),
	}
	return labelFilter, nil
}

// Create a labels.Requirement that can be used to build a filter
func (d *Deployment) LabelRequirementForPods() (*labels.Requirement, error) {
	return labels.NewRequirement(config.IdentLabelKey, selection.Equals, sets.NewString(d.identifier))
}

// Checks if the deployment is enrolled in kube-monkey
func (d *Deployment) IsEnrolled(client *kube.Clientset) (bool, error) {
	deployment, err := client.Extensions().Deployments(d.namespace).Get(d.name)
	if err != nil {
		return false, nil
	}
	return deployment.Labels[config.EnabledLabelKey] == config.EnabledLabelValue, nil
}

func (d * Deployment) HasKillAll(client *kube.Clientset) (bool, error) {
	deployment, err := client.Extensions().Deployments(d.namespace).Get(d.name)
	if err != nil {
		// Ran into some error: return 'false' for killAll to be safe
		return false, nil
	}

	return deployment.Labels[config.KillAllLabelKey] == config.KillAllLabelValue, nil
}

// Checks if this deployment is blacklisted
func (d *Deployment) IsBlacklisted(blacklist sets.String) bool {
	return blacklist.Has(d.namespace)
}
