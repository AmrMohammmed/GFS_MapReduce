package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type SlaveDevice struct {
	id     string
	ipAddr string
	data   string
}

var slave1 = SlaveDevice{id: "1", ipAddr: "http://192.168.1.5:8091"}
var slave2 = SlaveDevice{id: "2", ipAddr: "http://192.168.1.5:8092"}
var slave3 = SlaveDevice{id: "3", ipAddr: "http://192.168.1.5:8093"}
var slave4 = SlaveDevice{id: "4", ipAddr: "http://192.168.1.5:8094"}
var slave5 = SlaveDevice{id: "5", ipAddr: "http://192.168.1.5:8095"}
var Slaves = []SlaveDevice{slave1, slave2, slave3, slave4, slave5}
var reducer_Response = ""
var ResultChan = make(chan string)

func make_chunks(fileName string) {
	bytes, err := ioutil.ReadFile(fileName)
	panicOnError(err)
	// Allocate data to each slave device
	chunkSize := len(bytes) / 4

	slave1.data = "1" + string(bytes[:chunkSize])
	slave2.data = "2" + string(bytes[chunkSize:chunkSize*2])
	slave3.data = "3" + string(bytes[chunkSize*2:chunkSize*3])
	slave4.data = "4" + string(bytes[chunkSize*3:])

}
func divide_chunks_on_slaves() {

	make_chunks("Genome.fasta")

	Slaves = []SlaveDevice{slave1, slave2, slave3, slave4, slave5}

	for i := 0; i < 4; i++ {
		go func(SlavesArr []SlaveDevice, itr int) {
			b := strings.NewReader(SlavesArr[itr].data)
			resp, err := http.Post(SlavesArr[itr].ipAddr+"/fasta", "text/plain", b)
			panicOnError(err)
			defer resp.Body.Close()
			// get stautus code
			// fmt.Println("Status code:", resp.StatusCode)
			b = strings.NewReader(SlavesArr[itr+1].data)
			if itr == 3 {
				b = strings.NewReader(SlavesArr[0].data)
			}
			resp, err = http.Post(SlavesArr[itr].ipAddr+"/fasta", "text/plain", b)
			panicOnError(err)
			defer resp.Body.Close()
			// get stautus code
			ResultChan <- strconv.Itoa(resp.StatusCode)
		}(Slaves, i)

	}
	statusCodes := "\n"
	for i := 0; i < 4; i++ {
		statusCodes += (<-ResultChan + "\n")
	}
	fmt.Print("Status codes:", statusCodes)
	// os.Remove("Genome.fasta")

}

func get_from_slave(ipAdder string) (result_location string, IsLocation_flag bool) {

	resp, err := http.Get(ipAdder)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	// get stautus code
	fmt.Println("Status code:", resp.StatusCode)
	b, err := ioutil.ReadAll(resp.Body)
	result_location = string(b)
	IsLocation_flag = strings.Contains(result_location, "http://")

	return result_location, IsLocation_flag
}

func loop_over_get_requests(link string, CID string) string {
	var slaveIndex int
	var loopEnder int

	switch id := CID; id {
	case "0":
		slaveIndex = 0
		loopEnder = 4
	case "1":
		slaveIndex = 0
	case "2":
		slaveIndex = 1
	case "3":
		slaveIndex = 2
	case "4":
		slaveIndex = 3
	default:
		fmt.Println("Invalid id only ids from 0 to 4 is premitted")
		return ""
	}
	if CID != "0" {
		loopEnder = slaveIndex + 1
	}
	//var Slaves = []SlaveDevice{slave1, slave2, slave3, slave4, slave5}
	//indexes of slaves from the array above that 's what the loop of get requests  useses below

	FinalResult := ""

	for i := slaveIndex; i < loopEnder; i++ {

		go func(itr int, linker string) {
			resultLocation, Is_location_flag := get_from_slave(Slaves[itr].ipAddr + linker + strconv.Itoa(itr+1))

			if Is_location_flag == false && itr == 0 {
				resultLocation, _ = get_from_slave(slave4.ipAddr + linker + "1")
			} else if Is_location_flag == false {
				resultLocation, _ = get_from_slave(Slaves[itr-1].ipAddr + linker + strconv.Itoa(itr+1))
			}

			ResultChan <- resultLocation
		}(i, link)
	}
	for i := slaveIndex; i < loopEnder; i++ {

		FinalResult += (<-ResultChan + "\n")
	}

	FinalResult = FinalResult[:len(FinalResult)-1]
	fmt.Print(FinalResult + "\n")
	return FinalResult
}

func map_request(MapID string) {

	fmt.Println("Map and get Map Results from Slaves")

	mapResult := loop_over_get_requests("/chunk/baseCount?id=", MapID)

	//if reducer responded returns true else get out of the function

	if MapID == "0" {
		reducerResponsed := send_mapResult_to_reducer(slave5.ipAddr, mapResult)
		if reducerResponsed == false {
			reducerResponsed = send_mapResult_to_reducer(slave1.ipAddr, mapResult)
		}
	} else {

		//actually it should be named now mapper response because this the result of mapping of one slave but no need to new variable
		reducer_Response = mapResult
	}

}

func send_mapResult_to_reducer(ipAdder string, mapResult string) (reducerResponsed bool) {

	b := strings.NewReader(mapResult)
	resp, err := http.Post(ipAdder+"/reduceResults", "text/plain", b)
	//don't close the program if there is no response
	if err != nil {
		return
	}

	panicOnError(err)
	defer resp.Body.Close()
	// get stautus code
	fmt.Println("Status code:", resp.StatusCode)
	return true
}

// master acts as client  above
// ------------------------------------main is here !! -------------------------------------------------------------
func main() {

	// divide_chunks_on_slaves()
	master_as_server()
}

//master acts server below

func master_as_server() {
	http.HandleFunc("/", indexM)
	http.HandleFunc("/fasta", get_Slave_ip)
	http.HandleFunc("/fasta/baseCount", get_finalReducer_ip)
	http.HandleFunc("/reducerResponse", handle_reducer_response)

	fmt.Println("starting server")
	http.ListenAndServe(":8090", nil)

}

func indexM(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Handling / req")
	fmt.Fprintf(w, "Hello from Master")
}
func get_Slave_ip(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Handling GET req")
	// http://localhost:8090/fasta?id=0
	query := req.URL.Query()
	id := query.Get("id")
	var resultLocation string
	var err error

	//get_from_slave only gets a result location of fasta chunk not mapping or anything only mapping in maprequest function
	if id != "" {

		resultLocation = loop_over_get_requests("?id=", id)
		// set header return data
		w.Header().Add("Content-Type", "text/plain")
		fmt.Fprintf(w, id+"\n"+resultLocation)

	} else {
		resultLocation = "Write id value from 0 to 4 ,example:(http://localhost:8090/fasta?id=1)"
		fmt.Fprintf(w, resultLocation)
	}
	// if we had any error return status 500 and error
	if err != nil {
		errorHandlerM(w, req, http.StatusInternalServerError, err)
		return
	}

}

// listens to client
func get_finalReducer_ip(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Handling GET req")
	query := req.URL.Query()
	id := query.Get("id")
	map_request(id) //also map result is sent to reducer here if passed the constraints

	if reducer_Response != "" {
		fmt.Fprintf(w, id+"$mapReduceResult\n"+reducer_Response)

	} else {
		fmt.Fprintf(w, "Reducer or Mapper didn't respond maybe network failed\nor invalid id (note: vaild ids starts from 0 to 4)")
	}

}

func handle_reducer_response(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Handling POST req")
	defer req.Body.Close()

	// read req body
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		errorHandlerM(w, req, http.StatusInternalServerError, err)
		return
	}
	reducer_Response = string(b)
	w.WriteHeader(http.StatusCreated)
}

func panicOnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func errorHandlerM(w http.ResponseWriter, req *http.Request, status int, err error) {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "text/plain")
	fmt.Fprintf(w, `{error:%v}`, err.Error())
}
