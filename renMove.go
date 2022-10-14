package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var path = flag.String("path", ".", "working directory")
var flows = flag.Int("f", 1, "number of parallel flows")
var rename = flag.Bool("r", false, "rename files")

func main() {
	usage := "Use flag -path to point to working directory \n" +
		"Flag -f run job in different thread (more 4 not recommended) \n" +
		"If you want to rename files in mm-hh-ss format use flag -r "
	fmt.Println(usage)
	start := time.Now()
	num := 0
	flag.Parse()
	if *flows < 1 {
		*flows = 1
	}

	files, err := os.ReadDir(*path)
	if err != nil {
		fmt.Println("some error listing files")
	}
	if len(files) > 0 {
		fmt.Printf("Found %d files\n", len(files))
	} else {
		fmt.Printf("No files found")
		os.Exit(1)
	}
	step := len(files) / *flows
	n := make(chan int)
	for i := 0; i < *flows; i++ {
		if i+1 == *flows {
			go renMove(files[i*step:], n)
		} else {
			go renMove(files[i*step:(i+1)*step], n)
		}
	}
	for i := 0; i < *flows; i++ {
		num += <-n
	}
	fmt.Printf("\nmoved %d files\n", num)

	fmt.Printf("working time %d ms", time.Since(start).Milliseconds())
}

func renMove(files []os.DirEntry, n chan int) {
	num := 0
	k := 0
	for i, file := range files {
		if i > k*len(files)/60**flows {
			k++
			fmt.Print("+")
		}
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(*path, file.Name())
		var newPath string

		if info, _ := file.Info(); info.Size() == 0 {
			if err := os.Remove(filePath); err == nil {
				fmt.Println("Empty file removed")
			}
			continue
		}

		datetime, err := strToDate(file.Name())
		if err != nil {
			continue
		}
		year := datetime.Year()
		month := datetime.Month()
		day := datetime.Day()
		hour := datetime.Hour()
		minute := datetime.Minute()
		second := datetime.Second()
		if *rename {
			newPath = filepath.Join(*path, strconv.Itoa(year), month.String(), strconv.Itoa(day),
				fmt.Sprintf("%02d-%02d-%02d%s", hour, minute, second, filepath.Ext(file.Name())))
		} else {
			newPath = filepath.Join(*path, strconv.Itoa(year), month.String(), strconv.Itoa(day), file.Name())
		}

		if moveFile(filePath, newPath) != nil {
			continue
		}
		num++

	}
	n <- num
}

func moveFile(oldPath, newPath string) error {
	if _, err := os.Stat(filepath.Dir(newPath)); err != nil {
		if err := os.MkdirAll(filepath.Dir(newPath), 0777); err != nil {
			return err
		}
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	return nil
}

func strToDate(s string) (time.Time, error) {
	datetime, err := strconv.ParseInt(strings.TrimSuffix(s, filepath.Ext(s)), 10, 64)
	if err != nil {
		return time.Now(), err
	}
	return time.Unix(datetime, 0), nil

}
