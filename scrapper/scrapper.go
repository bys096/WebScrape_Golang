package scrapper

import (
	"encoding/csv"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type extractedJob struct {
	id       string
	title    string
	location string
	skills   string
}

// var baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?&searchword=python"

// Scrape SaramIn by a term
func Scrape(term string) {
	var baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?=&searchword=" + term + "&recruitSort=relation&recruitPageCount=40&inner_com_type=&company_cd=0%2C1%2C2%2C3%2C4%2C5%2C6%2C7%2C9%2C10&show_applied=&quick_apply=&except_read=&ai_head_hunting=&mainSearch=n"
	var jobs []extractedJob
	c := make(chan []extractedJob)
	totalPages := getPages(baseURL)

	for i := 1; i <= totalPages; i++ {
		go getPage(i, baseURL, c) // 1개 페이지의 모든 card 정보들이 배열로 담겨져있음
	}

	for i := 0; i < totalPages; i++ {
		extractedJobs := <-c
		jobs = append(jobs, extractedJobs...) // 그냥 append를 하면 이중배열이 됨으로, slice를 풀고 append
	}

	writeJobs(jobs)
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"ID", "Title", "Location", "Salary", "Summary"}
	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs {
		jobSlice := []string{job.id, job.title, job.location, job.skills}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}
}

func getPage(page int, baseURL string, mainC chan<- []extractedJob) {
	var jobs []extractedJob
	c := make(chan extractedJob)
	pageUrl := baseURL + "&recruitPage=" + strconv.Itoa(page)
	fmt.Println("Request", pageUrl)
	res, err := http.Get(pageUrl)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".content").Children()

	// card 마다 extractJob 생성하고, jobs에 추가
	searchCards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)
	})

	for i := 0; i < searchCards.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}

	mainC <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	id, _ := card.Attr("value")
	title := card.Find(".job_tit>a").Text()
	location := card.Find(".job_condition>span:first-child").Text()

	skillTags := []string{}
	card.Find(".job_sector>a").Each(func(i int, tag *goquery.Selection) {
		skillTags = append(skillTags, tag.Text())
	})
	skills := strings.Join(skillTags, " ")

	c <- extractedJob{
		id:       id,
		title:    title,
		location: location,
		skills:   skills,
	}
	//return extractedJob{
	//	id:       id,
	//	title:    title,
	//	location: location,
	//	skills:   skills,
	//}
}

func getPages(baseURL string) int {
	pages := 0
	res, err := http.Get(baseURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})

	return pages

	//fmt.Println(doc)

}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status", res.StatusCode)
	}
}
