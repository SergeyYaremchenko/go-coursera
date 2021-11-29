package main

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
)

func SingleHash(in, out chan interface{}) {
	fmt.Printf("%v - %v SingleHash start\n", in, out)
	outmd5 := make(chan string, 100)
	outcrc32 := make(chan string, 100)

	defer close(outcrc32)
	defer close(outmd5)

	for inData := range in {

		inDataStr := strconv.Itoa(inData.(int))

		fmt.Printf("%v - %v SingleHash[%s] data %s\n", in, out, inDataStr, inDataStr)

		md5 := DataSignerMd5(inDataStr)

		fmt.Printf("%v - %v SingleHash[%s] md5(data) %s\n", in, out, inDataStr, md5)

		go signCrc32Chan(md5, outmd5)

		go signCrc32Chan(inDataStr, outcrc32)

		crc32md5 := <-outmd5

		fmt.Printf("%v - %v SingleHash[%s] crc32(md5(data)) %s\n", in, out, inDataStr, crc32md5)

		crc32 := <-outcrc32

		fmt.Printf("%v - %v SingleHash[%s] crc32(data) %s\n", in, out, inDataStr, crc32)

		result := crc32 + "~" + crc32md5

		fmt.Printf("%v - %v SingleHash[%s] result %s\n", in, out, inDataStr, result)

		out <- result

		runtime.Gosched()
	}

	fmt.Printf("%v - %v SingleHash finish\n", in, out)
}

func signCrc32Chan(data string, out chan string) {
	crc32 := DataSignerCrc32(data)

	out <- crc32
}

func signCrc32Ptr(data string, out *string, wg *sync.WaitGroup) {
	crc32 := DataSignerCrc32(data)

	*out = crc32

	fmt.Printf("%s MultiHash: crc32(th+step1)) %s\n", data, crc32)

	wg.Done()
}

func MultiHash(in, out chan interface{}) {
	fmt.Printf("%v - %v MultiHash start\n", in, out)

	for inData := range in {
		data := inData.(string)
		go mhRoutine(data, out)
	}

	fmt.Printf("%v - %v MultiHash finish\n", in, out)
}

func mhRoutine(data string, out chan interface{}) {
	r := [6]string{"0", "1", "2", "3", "4", "5"}
	outData := make([]string, 6)
	wg := sync.WaitGroup{}
	for i, v := range r {
		go signCrc32Ptr(v+data, &outData[i], &wg)
		wg.Add(1)
	}

	wg.Wait()
	fmt.Printf("%s MultiHash result:%s\n", data, outData)
	result := ""

	for _, v := range outData {
		result += v
	}

	out <- result
}

func CombineResults(in, out chan interface{}) {
	fmt.Printf("%v - %v CombineResults start\n", in, out)
	inputData := make([]string, 0)
	for inDataUntyped := range in {
		inData := inDataUntyped.(string)
		fmt.Printf("%v - %v CombineResults received data[%s]\n", in, out, inData)
		inputData = append(inputData, inData)
	}

	if len(inputData) == 0 {
		return
	}
	sort.Strings(inputData)
	result := inputData[0]

	for i := 1; i < len(inputData); i++ {
		result += "_" + inputData[i]
	}
	fmt.Printf("%v CombineResults:%s\n", in, result)
	out <- result
	fmt.Printf("%v - %v CombineResults finish\n", in, out)
}

func wrapJob(j job, in, out chan interface{}, wg *sync.WaitGroup) {
	defer close(out)
	defer wg.Done()
	fmt.Printf("%v - %v job start\n", in, out)
	j(in, out)
	fmt.Printf("%v - %v job finish, closing %v\n", in, out, out)
}

func ExecutePipeline(jobs ...job) {
	wg := sync.WaitGroup{}
	in := make(chan interface{})
	out := make(chan interface{}, 100)

	for _, j := range jobs {
		wg.Add(1)
		go wrapJob(j, in, out, &wg)

		in = out
		out = make(chan interface{}, 100)
	}
	wg.Wait()
}

func main() {
	runtime.GOMAXPROCS(8)
	if len(os.Args) < 2 || os.Args[1] == "" {
		panic("Empty input")
	}
}
