package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gookit/color"
)

var BiaoJi []string
var w sync.WaitGroup

// Scans 单个扫描
func Scans(Turl string) {

	//状态码判断是否输入错误
	CodeIstrue(Codel(Rcode))
	//读取目录文件
	dic := Typeselection()
	//设置进度条
	var bar Bar
	bar.NewBar(0, len(dic))
	//设置字典管道
	pathChan := make(chan string, len(dic))
	//生产者
	for _, v := range dic {
		pathChan <- v
	}
	close(pathChan)

	//定时器
	if ProxyFile != "" {
		//每2秒随机切换代理
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			for range ticker.C {
				NewProxy = Randomget(ReadFile(ProxyFile), 1)
			}
		}()
	}

	//设置线程阻塞
	w.Add(Threads)
	//消费者
	for i := 0; i < Threads; i++ {
		//go func(url string,pathChan chan string,w sync.WaitGroup) {
		if Requestmode == "GET" {
			//fmt.Println("a:",runtime.NumGoroutine())
			go GetScan(Turl, pathChan, &w, &bar)
		} else if Requestmode == "HEAD" {
			go HeadScan(Turl, pathChan, &w, &bar)
		} else {
			fmt.Println("[!] 请输入正确的请求方式！")
			os.Exit(0)
		}
		//}(url,pathChan,w)
	}
	w.Wait()

	//递归扫描
	if Recursion == true {
		if len(BiaoJi) > 0 {
			//数组去重
			BiaoJi = RemoveRepByLoop(BiaoJi)
			//fmt.Println(BiaoJi)
			//进行递归扫描
			newurl := Urll(BiaoJi[0])
			//删除数据得第一个元素
			BiaoJi = BiaoJi[1:]
			fmt.Println(" ")
			color.Green.Printf("Target: %v \n", newurl)
			time.Sleep(200 * time.Millisecond)
			Scans(newurl)
		} else {
			color.Red.Printf("\n[!] 递归扫描结束")
		}
	}
	//bar.Close()
}

// Scanes 批量扫描
func Scanes() {
	//读取url列表
	urls := ReadFile(Urlfile)

	//遍历url
	for _, surl := range urls {
		Turl := Urll(surl)
		if FindUrl(Turl) == true {
			//fmt.Printf("\rtarget: %v\n",Turl)
			color.Green.Printf("\rtarget: %v\n", Turl)
			Scans(Turl)
		} else {
			//fmt.Printf("\rtarget: %v  [!] 目标url无法访问\n",Turl)
			color.Red.Printf("\rtarget: %v  [!] 目标url无法访问\n", Turl)
		}
	}

}

// AntiScans 反递归扫描
func AntiScans(aurl string) {
	//读取url列表
	urls := FDGtool(aurl)

	//遍历url
	for _, Aurl := range urls {
		Turl := Urll(Aurl)
		if FindUrl(Turl) == true {
			//fmt.Printf("\rtarget: %v\n",Turl)
			color.Green.Printf("\rtarget: %v\n", Turl)
			Scans(Turl)
		} else {
			//fmt.Printf("\rtarget: %v  [!] 目标url无法访问\n",Turl)
			color.Red.Printf("\rtarget: %v  [!] 目标url无法访问\n", Turl)
		}
	}

}

// HeadScan Scan Head扫描
func HeadScan(Turl string, pathChan <-chan string, w *sync.WaitGroup, bar *Bar) {
	for path := range pathChan {
		Targeturl := Turl + strings.Replace(path, "%", "%25", -1)
		resp := Request(Targeturl)
		if resp != nil {
			Rurl := resp.Header.Get("location") //获取302跳转的url
			respCode := resp.StatusCode         //状态码

			//指定状态码排除
			codes := Codel(Rcode)
			nocodes := Codel(Neglect)
			newcodes := difference(codes, nocodes)
			for _, code := range newcodes {
				if respCode == code {
					HeadPrint(respCode, Turl, path, Rurl)
					if Recursion == true {
						Recursionchoose(respCode, Turl, path)
					}
				}
			}

		}
		//进度条计数
		bar.Add(1)

	}
	w.Done()

}

// GetScan Getscan  Get扫描
func GetScan(Turl string, pathChan <-chan string, w *sync.WaitGroup, bar *Bar) {
	for path := range pathChan {
		Targeturl := Turl + strings.Replace(path, "%", "%25", -1)
		//Targeturl := Turl + path
		//Targeturl := Turl +`/`+ url.QueryEscape(path)
		//fmt.Println(Targeturl)
		resp := Request(Targeturl)
		if resp != nil {
			Rurl := resp.Header.Get("location") //获取302跳转的url
			body, _ := ioutil.ReadAll(resp.Body)
			Bodylen := Storage(len(body)) //返回长度
			//fmt.Println(Targeturl)
			//fmt.Println(string(body))
			respCode := resp.StatusCode //状态码

			//指定状态码排除
			codes := Codel(Rcode)
			nocodes := Codel(Neglect)
			newcodes := difference(codes, nocodes)
			for _, code := range newcodes {
				if respCode == code {
					//fmt.Println(Targeturl)
					GetPrint(respCode, Bodylen, body, Turl, path, Rurl)

					//记录递归扫描的目录
					if Recursion == true {
						Recursionchoose(respCode, Turl, path)
					}
				}
			}

		}
		//进度条计数
		bar.Add(1)
	}
	// 消费完毕则调用 Done，减少需要等待的线程
	w.Done()
}

// Processchecks 不进行随机代理时，每5秒检查扫描目标的存活
func Processchecks(Turl string) {
	//不进行随机代理时，每5秒检查扫描目标的存活
	if ProxyFile == "" && Checksurvive == false {
		color.Red.Printf("\r[!] 已开启存活检查！\n")
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			for range ticker.C {
				if !FindUrl(Turl) {
					fmt.Printf("\n网站无法访问,疑似waf或网络不通!\n")
					os.Exit(0)
				}
			}
		}()
	} else if ProxyFile == "" && Checksurvive == true {
		color.Red.Printf("\r[*] 已关闭存活检查！\n")
	}

}
