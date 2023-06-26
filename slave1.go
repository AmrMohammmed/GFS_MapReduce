package main

import (
	// "encoding/json"
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {

	http.HandleFunc("/", index)
	http.HandleFunc("/fasta", handleFasta)
	http.HandleFunc("/chunk/baseCount", handle_bases_count)
	http.HandleFunc("/countResult", handle_count_Result)
	http.HandleFunc("/reduceResults", handle_reduce_Results)
	http.HandleFunc("/finalbaseCount", handle_final_base_count)

	fmt.Println("starting server")
	http.ListenAndServe(":8091", nil)
}

func index(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Handling Get / req")
	query := req.URL.Query()
	chunk_ID := query.Get("id")

	// read file line by line
	if chunk_ID == "1" {
		fmt.Fprintf(w, "http://192.168.1.5:8091/fasta?id=1")

	} else if chunk_ID == "2" {
		fmt.Fprintf(w, "http://192.168.1.5:8091/fasta?id=2")
	}

}

func handleFasta(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		// fmt.Fprintf(w,"Nothing to Get yet")
		get(w, req)
	} else if req.Method == "POST" {
		post(w, req)
	} else {
		fmt.Println("Handling invalid /users req")
		errorHandler(w, req, http.StatusMethodNotAllowed, fmt.Errorf("Invalid Method"))
	}
}

func get(w http.ResponseWriter, req *http.Request) {

	fmt.Println("Handling Get / req")
	query := req.URL.Query()
	chunk_ID := query.Get("id")
	Input_fileName := ""

	// read file line by line
	if chunk_ID == "1" {
		Input_fileName = "slave1.fasta"

	} else if chunk_ID == "2" {
		Input_fileName = "slave2_copy.fasta"
	}

	file_bytes, err := ioutil.ReadFile(Input_fileName)
	fmt.Fprintf(w, string(file_bytes))
	if err != nil {
		errorHandler(w, req, http.StatusInternalServerError, err)
		return
	}

}

func post(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Handling POST req")
	defer req.Body.Close()

	// read req body
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		errorHandler(w, req, http.StatusInternalServerError, err)
		return
	}
	if string(b[0]) == "1" {
		err = ioutil.WriteFile("slave1.fasta", b[1:], 0644)
	} else if string(b[0]) == "2" {
		err = ioutil.WriteFile("slave2_copy.fasta", b[1:], 0644)
	}

	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusCreated)
}
func handle_bases_count(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Handling Get / req")

	query := req.URL.Query()
	map_chunk_ID := query.Get("id")
	Input_fileName := ""
	output_fileName := ""

	// read file line by line
	if map_chunk_ID == "1" {
		Input_fileName = "slave1.fasta"
		output_fileName = "countResult1.txt"

	} else if map_chunk_ID == "2" {
		Input_fileName = "slave2_copy.fasta"
		output_fileName = "countResult2.txt"
	}

	f, err := os.Open(Input_fileName)
	panicOnError(err)
	defer f.Close()

	Nucleobases := map[string]int{"A": 0, "T": 0, "C": 0, "G": 0}

	s := bufio.NewScanner(f)
	itr := 0
	for s.Scan() {
		// fmt.Println(s.Text())

		if itr == 0 && map_chunk_ID != "2" {
			itr = 1
			continue
		}
		Nucleobases["A"] += strings.Count(s.Text(), "A")
		Nucleobases["T"] += strings.Count(s.Text(), "T")
		Nucleobases["C"] += strings.Count(s.Text(), "C")
		Nucleobases["G"] += strings.Count(s.Text(), "G")
	}
	err = s.Err()
	panicOnError(err)

	countResult := "A:" + strconv.Itoa(Nucleobases["A"]) + "\nT:" + strconv.Itoa(Nucleobases["T"]) +
		"\nC:" + strconv.Itoa(Nucleobases["C"]) + "\nG:" + strconv.Itoa(Nucleobases["G"])

	f, err = os.Create(output_fileName)
	panicOnError(err)
	defer f.Close()
	f.WriteString(countResult)

	my_count_Result_location := "http://192.168.1.5:8091/countResult?id=" + map_chunk_ID

	fmt.Fprintf(w, my_count_Result_location)
	if err != nil {
		errorHandler(w, req, http.StatusInternalServerError, err)
		return
	}
}
func handle_count_Result(w http.ResponseWriter, req *http.Request) {

	fileName := ""
	query := req.URL.Query()
	countResult_ID := query.Get("id")
	if countResult_ID == "1" {
		fileName = "countResult1.txt"
	} else if countResult_ID == "2" {
		fileName = "countResult2.txt"
	}

	file_bytes, err := ioutil.ReadFile(fileName)
	fmt.Println("Handling / req")
	fmt.Fprintf(w, string(file_bytes))
	if err != nil {
		errorHandler(w, req, http.StatusInternalServerError, err)
		return
	}
	//instead of above if statments
	// os.Remove("countResult" + countResult_ID + ".txt")
}

// handle reducer request
func handle_reduce_Results(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Handling POST req")
	defer req.Body.Close()

	// read req body
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		errorHandler(w, req, http.StatusInternalServerError, err)
		return
	}

	//get results from mapers (slaves) to reduce
	slaveIP := string(b)
	IPs := strings.Split(slaveIP, "\n")

	Nucleobases := map[string]int{"A": 0, "T": 0, "C": 0, "G": 0}

	for i := 0; i < len(IPs); i++ {
		fmt.Println("Get from Slave:" + IPs[i])
		resp, err := http.Get(IPs[i])
		panicOnError(err)
		defer resp.Body.Close()
		fmt.Println("Status code:", resp.StatusCode)
		b, err = ioutil.ReadAll(resp.Body)

		lines_of_Nucleobases := strings.Split(string(b), "\n")

		for j := 0; j < len(lines_of_Nucleobases); j++ {
			kv_list := strings.Split(lines_of_Nucleobases[j], ":")
			key := kv_list[0]
			value, _ := strconv.Atoi(kv_list[1])
			Nucleobases[key] += value
		}

	}

	reduceResult := "A:" + strconv.Itoa(Nucleobases["A"]) + "\nT:" + strconv.Itoa(Nucleobases["T"]) +
		"\nC:" + strconv.Itoa(Nucleobases["C"]) + "\nG:" + strconv.Itoa(Nucleobases["G"])

	f, err := os.Create("final_fasta_baseCount.txt")
	panicOnError(err)
	defer f.Close()
	f.WriteString(reduceResult)

	if err != nil {
		errorHandler(w, req, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	send_reduceResult_location_to_master()
}

func send_reduceResult_location_to_master() {

	b := strings.NewReader("http://192.168.1.5:8091/finalbaseCount")
	resp, err := http.Post("http://192.168.1.5:8090/reducerResponse", "text/plain", b)
	panicOnError(err)
	defer resp.Body.Close()
	// get stautus code
	fmt.Println("Status code of send reduce result:", resp.StatusCode)
}

// handle get rq from client
func handle_final_base_count(w http.ResponseWriter, req *http.Request) {
	file_bytes, err := ioutil.ReadFile("final_fasta_baseCount.txt")
	fmt.Println("Handling / req")
	fmt.Fprintf(w, string(file_bytes))
	if err != nil {
		errorHandler(w, req, http.StatusInternalServerError, err)
		return
	}

}

func errorHandler(w http.ResponseWriter, req *http.Request, status int, err error) {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "text/plain")
	fmt.Fprintf(w, `{error:%v}`, err.Error())
}

func panicOnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
