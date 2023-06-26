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
	http.ListenAndServe(":8095", nil)
}

func index(w http.ResponseWriter, req *http.Request) {
	file_bytes, err := ioutil.ReadFile("slave5.fasta")
	fmt.Println("Handling / req")
	fmt.Fprintf(w, string(file_bytes))
	if err != nil {
		errorHandler(w, req, http.StatusInternalServerError, err)
		return
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
	file_bytes, err := ioutil.ReadFile("slave5.fasta")
	fmt.Println("Handling / req")
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

	err = ioutil.WriteFile("slave5.fasta", b, 0644)
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusCreated)
}

func handle_bases_count(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Handling / req")

	Nucleobases := map[string]int{"A": 0, "T": 0, "C": 0, "G": 0}

	// read file line by line
	f, err := os.Open("slave5.fasta")
	panicOnError(err)
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		// fmt.Println(s.Text())
		Nucleobases["A"] += strings.Count(s.Text(), "A")
		Nucleobases["T"] += strings.Count(s.Text(), "T")
		Nucleobases["C"] += strings.Count(s.Text(), "C")
		Nucleobases["G"] += strings.Count(s.Text(), "G")

	}
	err = s.Err()
	panicOnError(err)

	countResult := "A:" + strconv.Itoa(Nucleobases["A"]) + "\nT:" + strconv.Itoa(Nucleobases["T"]) +
		"\nC:" + strconv.Itoa(Nucleobases["C"]) + "\nG:" + strconv.Itoa(Nucleobases["G"])

	f, err = os.Create("countResult5.txt")
	panicOnError(err)
	defer f.Close()
	f.WriteString(countResult)

	my_count_Result_location := "http://192.168.1.5:8095/countResult"
	fmt.Fprintf(w, my_count_Result_location)
	if err != nil {
		errorHandler(w, req, http.StatusInternalServerError, err)
		return
	}
}

// respond to reducer
func handle_count_Result(w http.ResponseWriter, req *http.Request) {
	file_bytes, err := ioutil.ReadFile("countResult5.txt")
	fmt.Println("Handling / req")
	fmt.Fprintf(w, string(file_bytes))
	if err != nil {
		errorHandler(w, req, http.StatusInternalServerError, err)
		return
	}
	//instead of above if statments
	//os.Remove("countResult5.txt")
}

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

	b := strings.NewReader("http://192.168.1.5:8095/finalbaseCount")
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
