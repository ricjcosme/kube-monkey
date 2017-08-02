package healthandmetrics

import (
	"log"
	"net/http"
	"github.com/ricjcosme/kube-monkey/config"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"github.com/ricjcosme/kube-monkey/kubernetes"
	"k8s.io/client-go/1.5/pkg/api"
	"strconv"
	"time"
	"strings"
)

// TODO
// more possible metrics:
// number of pods killed by namespace
// number of times kube-monkey ran during daily start / end hours
// next run schedule


func Run() {
	// Handler for returning health check on /healthz endpoint
	// Started in main as a go routine
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("ok"))

	})
	http.HandleFunc("/chaosmetrics", func(w http.ResponseWriter, r *http.Request) {

		// Retrieve all K8s events created by kube-monkey
		evs, err := KubeMonkeyEvents()

		// on error return HTTP 500
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		}

		// Get the counters
		nMin, nHour, nDay := LatestEvents(evs)
		var metricsBody = []string{"pods_killed_last_5_min ", strconv.Itoa(nMin), "\npods_killed_last_60_min ",
			strconv.Itoa(nHour), "\npods_killed_last_24_hours ", strconv.Itoa(nDay)}
		// Return HTTP 200 and the body
		w.WriteHeader(200)
		w.Write([]byte(strings.Join(metricsBody, " ")))
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// KubeMonkeyEvents retrieve all K8s events
// created by kube-monkey
func KubeMonkeyEvents() ([]v1.Event, error) {
	client, err := kubernetes.NewInClusterClient()
	if err != nil {
		return nil, err
	}

	allEvents, err := client.Events(config.WhitelistedNamespaces).List(api.ListOptions{})
	if err != nil {
		return nil, err
	}

	kubeMonkeyEvents := []v1.Event{}

	for _, event := range allEvents.Items {
		if event.Source.Component == config.KubeMonkeyAppName() {
			kubeMonkeyEvents = append(kubeMonkeyEvents, event)
		}
	}

	return kubeMonkeyEvents, nil
}

// LatestEvents counts the number of pods killed
// in the latest 5 minutes, 1 hour and 1 day
// and returns those values
func LatestEvents(eventList []v1.Event) (countMinutes int, countHour int, countDay int) {
	var cMin, cHour, cDay int = 0, 0, 0
	var elapsedTime5 float64 = 5
	for _, event := range eventList {
		if time.Since(event.LastTimestamp.Time).Minutes() <= elapsedTime5 {
			cMin++
		}
		if time.Since(event.LastTimestamp.Time).Minutes() <= (elapsedTime5*12) {
			cHour++
		}
		if time.Since(event.LastTimestamp.Time).Minutes() <= (elapsedTime5*12*24) {
			cDay++
		}
	}
	return cMin, cHour, cDay
}