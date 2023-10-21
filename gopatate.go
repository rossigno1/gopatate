package main

import (
	"github.com/rossigno1/requests"
	"net/http"
	"strings"
	"log"
	"fmt"
	"bufio"
	"regexp"
	"flag"
	"strconv"
	"sync"
	"os"
	"io/ioutil"
)

var (
 RESET  = "\033[0m"
 RED    = "\033[31m"
 GREEN  = "\033[32m"
 YELLOW = "\033[33m"
 Blue   = "\033[34m"
 Purple = "\033[35m"
 CYAN   = "\033[36m"
 Gray   = "\033[37m"
 White  = "\033[97m"
)

type config struct {
	Url string 
	Method string
	MinSize int64 
	MaxSize int64 
	Extensions []string
	Codes []int
	Data []byte
	Headers map[string]string
	OutSize int64
	Msg string
	Threads int
}



func (conf *config)SearchInUrl(files []string,logfile string){

	var reg *regexp.Regexp //for content check 
	isget := true //for method

	checkcontent := false
	if len(conf.Msg) > 0 {
		checkcontent = true
		reg = regexp.MustCompile(conf.Msg)
	}

	logs,err := os.OpenFile(logfile,os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err !=nil { return }
	defer logs.Close()
	//Threading 
	wg := sync.WaitGroup{}

	wg.Add(conf.Threads)
	chPage := make(chan string, conf.Threads)

	if strings.ToLower(conf.Method) == "head" {
		isget = false
	}

	for i := 0; i < conf.Threads; i++ {
		go func(ch chan string, tid int) {
			defer wg.Done()
			for u := range ch {
				path := strings.Join(strings.Split(u,"/")[3:],"/")

				var r *http.Response
				if isget {
					r,err = requests.Get(u,nil,conf.Headers,nil,false)
				}else{
					r,err = requests.Head(u,nil,conf.Headers,nil,false)
				}
				if err != nil { log.Printf(" XX | %s - %s \n",path,err);continue}
				defer r.Body.Close()
				nop := false 
				for _,c := range conf.Codes {
					if r.StatusCode == c {
						nop = true 
						break
					}
				}
				if nop == false {
					body,_ := ioutil.ReadAll(r.Body)
					bsize:= int64(len(body))
					if conf.OutSize == bsize {continue}
					if conf.MinSize > 0 && conf.MaxSize > 0 {
						if bsize >= conf.MinSize && bsize <= conf.MaxSize  {continue }
					}
					if checkcontent {
						if !reg.Match(body) { continue }
					}
					status := ""
					switch r.StatusCode {
						case 200 : status = fmt.Sprintf("%s %d %s",GREEN,r.StatusCode,RESET)
						case 301 : status = fmt.Sprintf("%s %d %s",YELLOW,r.StatusCode,RESET)
						case 302 : status = fmt.Sprintf("%s %d %s",CYAN,r.StatusCode,RESET)
						case 401 : fallthrough 
						case 403 : status = fmt.Sprintf("%s %d %s",RED,r.StatusCode,RESET)
						default : status = fmt.Sprintf("%d",r.StatusCode)
					}
					log.Printf("[%s] - %s , %d\n",status, path, bsize)
					logs.Write([]byte(fmt.Sprintf("%d,%s,%s,%d\n",r.StatusCode,u,path,bsize)))
				}
				
			}
		}(chPage, i)
	}

	//Just wirh one file 
	for _,f := range files {

		//Get file size for progressbarr 
		d,err := os.ReadFile(f)
		if err != nil { log.Println(f,err);continue}
		total := len(strings.Split(string(d),"\n"))
		d = []byte{}

		//Progress bar
		pwg := sync.WaitGroup{}
		pwg.Add(1)
		chProgress:= make(chan int, 1)
		go func(ch chan int) {
			wg.Done()
			idx := 0
			for i := range ch {
				idx+=i
				fmt.Printf("[%d/%d]\r",idx,total)
			}

		}(chProgress)

		//Open scanner
		file, err := os.Open(f)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		// optionally, resize scanner's capacity for lines over 64K, see next example
		for scanner.Scan() {
			u := conf.Url 
			u = strings.Replace(u,"FILE0",scanner.Text(),1)
			chPage <- u
			chProgress <- 1
			for _,ext := range conf.Extensions {
				if len(ext) > 1 {
					ue := fmt.Sprintf("%s.%s",u,ext)
					chPage <- ue
				}
			}
			
		}
		close(chProgress)
		pwg.Wait()
	}
	close(chPage)
	wg.Wait()
}

func main(){
	var (
		conf config 
		filename string 
		ext string
		exclude string
		cookies string
		heads string
		outlog string
		data string
	)
	flag.StringVar(&conf.Url,"u","","URL to scan with FILE0 for the place of the iteration ")
	flag.StringVar(&filename,"f","","Files used for iteration separated by comma. First file is for FILE0 second is for FILE1, etc.")
	flag.StringVar(&ext,"ext","","Extensions to use separate by coma php,asp,html,xml.")
	flag.StringVar(&exclude,"x","","Exclusion like code=404,size=1223-1227,msg=(.)*failed(.)*")
	flag.StringVar(&cookies,"cookie","","Cookie to add like PHPSESSID=AAAAAAAAA,lang=BBBBBB")
	flag.StringVar(&heads,"headers","","Headers to add like X-Forwarded-For:AAAAAAAAA,FOO:BBBBBB")
	flag.StringVar(&outlog,"o","patalog.csv","Logs the output")
	flag.StringVar(&conf.Method,"m","get","method use for brute force, only : get,head,post")
	flag.StringVar(&data,"d","","data to send with post")
	flag.IntVar(&conf.Threads,"t",5,"Threads")

	flag.Parse()

	if len(conf.Url) < 2 || len(filename) < 1 {
		log.Fatal("Missing url or filename")
	}


	files := strings.Split(filename,",")
	extensions := strings.Split(ext,",")
	conf.Headers = make(map[string]string)
	if len(heads) > 2 {
		for _,h := range strings.Split(heads,",") {
			head := strings.Split(h,":")
			conf.Headers[head[0]] = strings.Join(head[1:],":")
		}
	}

	//In what field do we search 
	regFile := regexp.MustCompile("FILE[0-9]+")
	search := "url"
	if regFile.MatchString(conf.Url) {
		search = "url"
	}
	if regFile.MatchString(heads) {
		search = "head"
	}

	exclusions := strings.Split(exclude,",")

	for _,e := range exclusions {
		if strings.Contains(e,"code=") {
			tmp := strings.Replace(e,"code=","",-1)
			code,err := strconv.Atoi(tmp)
			if err != nil { log.Fatal(e,err)}
			conf.Codes = append(conf.Codes,code)
			continue
		}
		if strings.Contains(e,"size=") {
			tmp := strings.Replace(e,"size=","",-1)
			t := strings.Split(tmp,"-")
			if len(t) == 1 {
				code,err := strconv.Atoi(t[0])
				if err != nil { log.Fatal(e,err)}
				conf.OutSize = int64(code) 
			}
			if len(t) > 1 {
				code,err := strconv.Atoi(t[0])
				if err != nil { log.Printf("Taking 0 as minsize");code = 0}
				conf.MinSize = int64(code)
				code,err = strconv.Atoi(t[1])
				if err != nil { log.Printf("Taking 9999999 as maxsize");code = 9999999}
				conf.MaxSize = int64(code)
			}
			continue
		}


		if strings.Contains(e,"msg=") {
			tmp := strings.Replace(e,"msg=","",-1)
			conf.Msg=tmp
			continue
		}
	}

	


	/********************* LOOK FOR URL **************/
	fmt.Printf(`Let's GOOOOOOOOO :
URL : %s
Headers : %v
Exclusions : %v
Extensions : %v
`,conf.Url,conf.Headers,conf.Codes,extensions)
	switch search {
		case "url" :conf.SearchInUrl(files,outlog)
	}

}