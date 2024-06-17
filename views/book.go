package views

import (
	"bufio"
	"errors"
	"go-reader/dao"
	"go-reader/utils"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding/simplifiedchinese"
)

type BookDir struct {
	name  string
	start int
}

var updateBookPosDebounce = utils.NewDebouncer(200)

var bookAll []string
var bookDirs []BookDir
var bookName string

func init() {
	bookAll = make([]string, 0)
	bookDirs = make([]BookDir, 0)
}

func ImportBook(filepath string) (err error) {
	all := make([]string, 0)
	cnt := 0
	// 读取txt文件
	file, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer file.Close()

	// /** 识别文件编码 **/
	reader := bufio.NewReader(file)
	b, err := reader.Peek(4096) // Peek at the first 1024 bytes
	if err != nil {
		return
	}
	// Detect the encoding
	result, err := chardet.NewTextDetector().DetectBest(b)
	if err != nil {
		return
	}
	useTrans := true
	// 判断result.Charset是否以GB开头
	if result.Charset[:2] == "GB" || result.Charset[:2] == "gb" {
		useTrans = true
	} else if result.Charset[:3] == "UTF" || result.Charset[:3] == "utf" {
		useTrans = false
	} else {
		err = errors.New("Unknown encoding: " + result.Charset)
		return
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var line string
		if useTrans {
			src := scanner.Text()
			buffer := make([]byte, len([]byte(src))*2)
			n, _, _ := simplifiedchinese.GB18030.NewDecoder().Transform(buffer, []byte(src), true)
			line = string(buffer[:n])
		} else {
			line = scanner.Text()
		}
		// 删除所有空行
		if line != "" {
			all = append(all, line)
			cnt++
		}
	}
	err = scanner.Err()
	if err != nil {
		return
	}

	// 写入数据库
	bookname, ext, _ := PathProc(filepath)
	_, err = dao.CreateBook(bookname, cnt)
	if err != nil {
		return
	}

	// 在当前目录下创建新的txt文件
	// 检查目录是否存在，如果不存在则创建
	_, err = os.Stat("download")
	if os.IsNotExist(err) {
		err = os.Mkdir("download", 0755)
		if err != nil {
			return
		}
	}
	newFile, err := os.Create("download/" + bookname + ext)
	if err != nil {
		return
	}
	defer newFile.Close()
	writer := bufio.NewWriter(newFile)
	for _, line := range all {
		writer.WriteString(line + "\n")
	}
	writer.Flush()
	return
}

func ProcBook(filename string) (bookLastPos int, err error) {
	book, err := dao.GetBookByName(filename)
	if err != nil {
		return
	}
	bookName = book.Title
	bookLastPos = book.LastPos
	path := "download/" + filename + ".txt"
	bookAll = make([]string, 0)
	bookDirs = make([]BookDir, 0)
	cnt := 0
	// 读取txt文件
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// /第[一二三四五六七八九十百千万零〇0-9]+(章|卷)/
	re := regexp.MustCompile(`第[一二三四五六七八九十百千万零〇0-9]+(章|卷)`)
	for scanner.Scan() {
		line := scanner.Text()
		bookAll = append(bookAll, line)
		if re.MatchString(line) {
			bookDirs = append(bookDirs, BookDir{name: line, start: cnt})
		}
		cnt++
	}
	err = scanner.Err()
	return
}

func GetBookContent(lastPos int) (title string, content string, bookCurrentIndex int) {
	if len(bookDirs) > 1 && lastPos < bookDirs[0].start {
		// 第0章，没有章节名,从0开始到第一章起点的前一行
		title, content = "", strings.Join(bookAll[0:bookDirs[0].start-1], "\n")
		bookCurrentIndex = -1
		return
	}
	for i := 0; i < len(bookDirs); i++ {
		if lastPos >= bookDirs[i].start {
			start := bookDirs[i].start
			if i == len(bookDirs)-1 {
				// 最后一章
				bookCurrentIndex = i
				title = bookDirs[i].name
				content = strings.Join(bookAll[start+1:], "\n")
				return
			}
			if lastPos < bookDirs[i+1].start {
				bookCurrentIndex = i
				title = bookDirs[i].name
				if start+1 >= bookDirs[i+1].start-1 {
					content = ""
					return
				}
				content = strings.Join(bookAll[start+1:bookDirs[i+1].start], "\n")
				return
			}
		}
	}
	return "", "", -1
}

func GetChapterStart(chapterIndex int) int {
	if chapterIndex < 0 {
		return 0
	}
	if chapterIndex >= len(bookDirs) {
		return len(bookAll)
	}
	return bookDirs[chapterIndex].start
}

func UpdateBookPos(name string, pos int) {
	// 防抖
	updateBookPosDebounce.Debounce(func() {
		dao.UpdateBookPos(name, pos)
	})
}

func DelBook(name string) error {
	err := dao.DeleteBookByName(name)
	if err != nil {
		return err
	}
	// 删除文件
	os.Remove("download/" + name + ".txt")
	return nil
}

func PathProc(path string) (name string, ext string, route string) {
	route = filepath.Dir(path)
	filenameWithExt := filepath.Base(path)
	ext = filepath.Ext(filenameWithExt)                       // 获取后缀
	name = filenameWithExt[0 : len(filenameWithExt)-len(ext)] // 去掉后缀
	return
}
