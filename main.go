package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	host     string
	ports    string
	threads  int
	timeout  int
	ping     bool
	filename string
	ipfile   string
	l        sync.RWMutex
)

func init() {
	flag.StringVar(&host, "h", "", "Target of scan task:192.168.0.1-192.168.0.255")
	flag.StringVar(&ports, "p", "21,22,23,25,53,80,110,139,143,389,443,445,465,873,993,995,1080,1723,1433,1521,3306,3389,3690,5432,5800,5900,6379,7001,8000,8001,8080,8081,8888,9200,9300,9080,9999,11211,27017", "The port need to be scan:80,81")
	flag.IntVar(&threads, "m", 100, "The threads of scan")
	flag.IntVar(&timeout, "t", 10, "Timeout num of each task")
	flag.BoolVar(&ping, "n", false, "Check the host is online")
	flag.StringVar(&filename, "f", "", "The input file name")
	flag.StringVar(&ipfile, "o", "", "The output file name")
	flag.Parse()

}

func worker(id int, jobs <-chan string) {
	for j := range jobs {
		scanner(j)
	}

}
func writeResult(data string) {
	filePath := ipfile
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Open file error", err)
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	write.WriteString(data + "\n")

	write.Flush()
}
func scanner(j string) {
	conn, err := net.DialTimeout("tcp", j, time.Second*9)
	if err == nil {

		fmt.Printf("%s is open\n", j)
		if ipfile != "" {
			l.RLock()
			writeResult(j + " is open")
			l.RUnlock()

		}
		defer conn.Close()
	}

}

func getPorts(param string) []string {
	var slice = []string{}
	if strings.Contains(param, ",") {
		portArr := strings.Split(param, ",")
		return portArr
	}
	if strings.Contains(param, "-") {
		ips := strings.Split(param, "-")
		startNum, _ := strconv.Atoi(ips[0])
		endNum, _ := strconv.Atoi(ips[1])
		for i := startNum; i <= endNum; i++ {
			slice = append(slice, strconv.Itoa(i))
		}
		return slice
	} else {
		slice = append(slice, param)
		return slice
	}

}

func stringSliceToInt(data []string) []int {
	ints := make([]int, len(data))
	for i, s := range data {

		ints[i], _ = strconv.Atoi(s)
	}
	return ints
}
func getIps(data string) []string {
	var slice = make([]string, 0)
	data = strings.Replace(data, "\\n", "", -1)
	startIp := strings.Split(strings.Split(data, "-")[0], ".")
	endIp := strings.Split(strings.Split(data, "-")[1], ".")
	startIpInt := stringSliceToInt(startIp)
	endIpInt := stringSliceToInt(endIp)

	if startIpInt[0] != endIpInt[0] {
		for i := startIpInt[0]; i <= endIpInt[0]; i++ {
			for j := 1; j <= 255; j++ {
				for k := 1; k <= 255; k++ {
					for t := 1; t <= 255; t++ {
						slice = append(slice, strconv.Itoa(i)+"."+strconv.Itoa(j)+"."+strconv.Itoa(k)+"."+strconv.Itoa(t))
					}

				}
			}

		}

	} else if startIpInt[1] != endIpInt[1] {
		for i := startIpInt[1]; i <= endIpInt[1]; i++ {
			for j := 1; j <= 255; j++ {
				for k := 1; k <= 255; k++ {
					slice = append(slice, strconv.Itoa(endIpInt[0])+"."+strconv.Itoa(i)+"."+strconv.Itoa(j)+"."+strconv.Itoa(k))
				}
			}

		}
	} else if startIpInt[2] != endIpInt[2] {
		for i := startIpInt[2]; i <= endIpInt[2]; i++ {
			for j := 1; j <= 255; j++ {
				slice = append(slice, strconv.Itoa(endIpInt[0])+"."+strconv.Itoa(endIpInt[1])+"."+strconv.Itoa(i)+"."+strconv.Itoa(j))
			}

		}
	} else if startIpInt[3] != endIpInt[3] {
		for i := startIpInt[3]; i <= endIpInt[3]; i++ {
			slice = append(slice, strconv.Itoa(startIpInt[0])+"."+strconv.Itoa(startIpInt[1])+"."+strconv.Itoa(startIpInt[2])+"."+strconv.Itoa(i))

		}
	}
	return slice

}
func getHosts(jobs chan string, ports []string) {

	if host == "" {
		if filename != "" {
			file, err := os.Open(filename)
			if err != nil {
				panic(err)
			}
			content, err := ioutil.ReadAll(file)
			lines := strings.Split(string(content), "\n")
			for _, ipstr := range lines {
				ipstr = strings.TrimSpace(ipstr)
				for _, ip := range getIps(ipstr) {
					for _, port := range ports {
						jobs <- ip + ":" + port
					}

				}

			}

		} else {
			log.Fatal("No target found")
			os.Exit(1)
		}

	} else if strings.Contains(host, "-") {

		for _, ip := range getIps(host) {

			for _, port := range ports {

				jobs <- ip + ":" + port
			}

		}

	} else {
		for _, port := range ports {
			jobs <- host + ":" + port
		}

	}

}

func main() {

	var wg sync.WaitGroup
	jobs := make(chan string, 300)
	portArr := getPorts(ports)
	wg.Add(1)
	go func() {
		getHosts(jobs, portArr)
		time.Sleep(time.Second * 30)
		close(jobs)

		defer wg.Done()
	}()

	for i := 1; i <= threads; i++ {
		wg.Add(1)
		go func(i int) {

			worker(i, jobs)
			defer wg.Done()
		}(i)

	}

	wg.Wait()

}
