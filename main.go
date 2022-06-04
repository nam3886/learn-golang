package main

import (
	"fmt"
	"runtime"
)

// chương trình crawl url với 5 goroutines
func main() {
	numberOfRequests := 100000
	maxWorkerNumber := 5
	queueChan := make(chan int, numberOfRequests) // buffer channel bản chất như queue
	doneChan := make(chan int)

	for i := 1; i <= maxWorkerNumber; i++ { // chạy 5 goroutines mỗi goroutine sẽ chạy vòng for để lấy dữ liệu từ channel ra để xử lý
		go func(workerName string) {
			for value := range queueChan { // lấy dữ liệu từ trong channel ra
				crawl(value, workerName)
			}

			doneChan <- 1
		}(fmt.Sprintf("%d", i)) // ở đây phải truyền tham số vào vì hàm này ko chạy liền (việc chạy do go runtime quản lý) nên biến i có thể đã bị thay đổi khi mà hàm này chạy => ko còn đúng nữa
	}

	for i := 1; i <= numberOfRequests; i++ { // đẩy dữ liệu vào channel
		queueChan <- i
	}

	// Người gửi có thể đóng một kênh để chỉ ra rằng sẽ không có thêm giá trị nào được gửi. => channel này chỉ có thể chứa numberOfRequests => for ở line 17 5 goroutines sẽ lấy ra đúng numberOfRequests và thoát for chứ ko phải chờ khi mà buffer channel empty
	close(queueChan) // nếu ko close channel => sẽ bị memory leak ở line 19 vì khi lấy dữ liệu từ empty buffer channel

	for i := 1; i <= maxWorkerNumber; i++ { // lấy dữ liệu từ doneChan ra khi doneChan (không phải buffer channel) empty => main goroutine không còn bị block => không cần dùng sleep để main không exit
		<-doneChan
	}

	// time.Sleep(time.Second * 5) // phải sleep ở đây vì ở trên các lệnh được đưa vào goroutines nếu ko sleep => exit luôn chương trình
}

func crawl(value int, name string) {
	// https://stackoverflow.com/questions/13107958/what-exactly-does-runtime-gosched-do
	// GOMAXPROCS nếu ko hiệu chỉnh => chỉ chạy trên 1 OS thread => ko thể switch execution contexts khi mà 1 goroutine đang thực thi
	runtime.Gosched() // nói với chương trình hãy switch execution contexts => chuyển sang 1 goroutine khác để thực thi [cách 1]
	// time.Sleep(time.Second / 100) // sleep ở đây để switch execution contexts (không 1 goroutine nào chạy liên tục) [cách 2]
	fmt.Println("Worker ", name, " is crawling: ", value)
}

// sync.Mutex
// count ++
// chương trình sẽ dịch thành:
// 1. read count (khi mà đọc đến đây goroutine đang bị giữa tiến trình => sẽ switch execution contexts)
// => sẽ có ít nhất 1 goroutine khác cũng đọc giá trị giống goroutine này => cả 2 goroutine sẽ cùng cộng lên trùng 1 giá trị
// => lỗi data racing (https://viblo.asia/p/007-data-race-va-mutual-exclusion-4dbZNGvmlYM)
// 2. add 1 to count
// 3. write count

// dùng sync.Mutex để fix data racing
// lock := new(sync.Mutex)
// lock.Lock() khi goroutine nào đến đây trước => acquire lock dù có switch execution contexts khi đang giữa tiến trình thì khi goroutine khác vào đây sẽ bị kẹt lại => goroutine nào giữ lock sẽ được thực hiện đếm lên và unlock rồi thực hiện lại bước trên
// => synchronize: trong 1 thời điểm chỉ có 1 goroutine thực hiện lệnh đếm lên
// count ++
// lock.Unlock()