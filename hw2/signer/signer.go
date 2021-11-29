package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
)

type dataDto struct {
	Dto         string
	InitialData string
}

func SingleHash(in, out chan interface{}) {
	fmt.Printf("%v - %v SingleHash start\n", in, out)

	outmd5 := make(chan string)
	outcrc32 := make(chan string)
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

		out <- dataDto{
			Dto:         result,
			InitialData: inDataStr,
		}
	}

	fmt.Printf("%v - %v SingleHash finish\n", in, out)
}

func signCrc32Chan(data string, out chan string) {
	crc32 := DataSignerCrc32(data)

	out <- crc32
}

func signCrc32Ptr(data string, initial string, inCh chan interface{}, outCh chan interface{}, out *string, wg *sync.WaitGroup) {
	fmt.Printf("%v - %v signCrc32Ptr[%s] got data: %s\n", inCh, outCh, initial, data)

	crc32 := DataSignerCrc32(data)

	*out = crc32

	fmt.Printf("%v - %v signCrc32Ptr[%s] result: %s\n", inCh, outCh, initial, crc32)

	wg.Done()
}

func MultiHash(in, out chan interface{}) {
	fmt.Printf("%v - %v MultiHash start\n", in, out)

	wg := sync.WaitGroup{}

	for inData := range in {
		data := inData.(dataDto)
		go mhRoutine(data, in, out, &wg)
		wg.Add(1)
	}

	wg.Wait()
	fmt.Printf("%v - %v MultiHash finish\n", in, out)
}

func mhRoutine(data dataDto, in, out chan interface{}, wg *sync.WaitGroup) {
	fmt.Printf("%v - %v mhRoutine[%s] got data: %v\n", in, out, data.InitialData, data.Dto)

	r := [6]string{"0", "1", "2", "3", "4", "5"}

	outData := make([]string, 6)
	wgr := sync.WaitGroup{}

	for i, v := range r {
		wgr.Add(1)

		go signCrc32Ptr(v+data.Dto, data.InitialData, in, out, &outData[i], &wgr)
	}

	wgr.Wait()

	fmt.Printf("%v - %v mhRoutine[%s] result: %v\n", in, out, data.InitialData, outData)
	result := ""

	for _, v := range outData {
		result += v
	}

	out <- result

	wg.Done()
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
	defer wg.Done()
	defer close(out)
	fmt.Printf("%v - %v job start\n", in, out)
	j(in, out)
	fmt.Printf("%v - %v job finish, closing %v\n", in, out, out)
}

func ExecutePipeline(jobs ...job) {
	wg := sync.WaitGroup{}
	in := make(chan interface{})
	out := make(chan interface{})

	for _, j := range jobs {
		wg.Add(1)
		go wrapJob(j, in, out, &wg)

		in = out
		out = make(chan interface{})
	}
	wg.Wait()
}

func main() {
	//runtime.GOMAXPROCS(8)
	if len(os.Args) < 2 || os.Args[1] == "" {
		panic("Empty input")
	}
}
