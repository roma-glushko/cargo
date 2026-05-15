package pathfinder

import (
	"math/rand"
	"time"
)

type TransitEdge struct {
	VoyageNumber string
	From         string
	To           string
	FromDate     time.Time
	ToDate       time.Time
}

type TransitPath struct {
	Edges []TransitEdge
}

func FindShortestPath(origin, destination string, allLocations []string, voyageNumbers []string) []TransitPath {
	var candidates []string
	for _, loc := range allLocations {
		if loc != origin && loc != destination {
			candidates = append(candidates, loc)
		}
	}

	numPaths := 3 + rand.Intn(3)
	var paths []TransitPath

	for i := 0; i < numPaths; i++ {
		numStops := 1 + rand.Intn(min(4, len(candidates)))
		shuffled := make([]string, len(candidates))
		copy(shuffled, candidates)
		rand.Shuffle(len(shuffled), func(a, b int) { shuffled[a], shuffled[b] = shuffled[b], shuffled[a] })

		stops := shuffled[:numStops]
		route := make([]string, 0, numStops+2)
		route = append(route, origin)
		route = append(route, stops...)
		route = append(route, destination)

		var edges []TransitEdge
		t := time.Now().Add(24 * time.Hour)

		for j := 0; j < len(route)-1; j++ {
			departure := t.Add(time.Duration(rand.Intn(12)) * time.Hour)
			arrival := departure.Add(24*time.Hour + time.Duration(rand.Intn(48))*time.Hour)

			vn := voyageNumbers[rand.Intn(len(voyageNumbers))]

			edges = append(edges, TransitEdge{
				VoyageNumber: vn,
				From:         route[j],
				To:           route[j+1],
				FromDate:     departure,
				ToDate:       arrival,
			})

			t = arrival
		}

		paths = append(paths, TransitPath{Edges: edges})
	}

	return paths
}
