package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

// Example-ish: {"imsi":"295050910000000","ussdDataCodingScheme":15,"ussdString":"*901031*1234567#","value":"1234567"}
type SoracomBeamData struct {
	Imsi string `json:"imsi"`
	CodingSheme int `json:"ussdDataCodingScheme"`
	UssdString string `json:"ussdString"`
	Value string `json:"value"`
}


func respondWithError(w http.ResponseWriter, code int, err error) {
	respondWithJSON(w, code, map[string]string{"error": err.Error()})
}

func respondWithStatus(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"status": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"error\": \"unable to build JSON response\"}"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}


func getSoracomAuthKey() string {
	return os.Getenv("SORACOM_BEAM_PSK")
}

func getWateringUpstream() string {
	return os.Getenv("WATERING_UPSTREAM")
}

func assembleWaterUrl(baseUrl string, seconds int) string {
	return fmt.Sprintf("%s/seconds/%d", baseUrl, seconds)
}

func WaterHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err)
		return
	}

	beamData := &SoracomBeamData{}
	err = json.Unmarshal(body, beamData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	log.Printf("Beam Data: %+v", beamData)

	activationSeconds, err := strconv.Atoi(beamData.Value)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	resp, err := http.Get(assembleWaterUrl(getWateringUpstream(), activationSeconds))
	if err != nil {
		statusCode := 503
		if resp != nil {
			statusCode = resp.StatusCode
		}
		respondWithError(w, statusCode, err)
		return
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	log.Println(string(respBody))

	respondWithStatus(w, http.StatusAccepted, "Thank You")
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/water/herbs", WaterHandler)

	log.Fatal(http.ListenAndServe(":8000", r))
}
