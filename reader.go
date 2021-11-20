package main

import (
	"os"
	"log"
	"strings"
	"unicode"
)

type Reader struct {
	data		string
	words		[]string
}


func NewReader() *Reader {
	return &Reader{}
}

func (self *Reader) Get(index int) string {
	if index < len(self.words) {
		return self.words[index]
	}
	return "- THE END -"
}

func (self *Reader) parseData() {

	var word strings.Builder

	for _, ch := range self.data {
//		log.Printf("ch = %q", ch)
		if unicode.IsLetter(ch) || unicode.IsNumber(ch) {
			word.WriteRune(ch)
		} else if unicode.IsPunct(ch) || unicode.IsSpace(ch) {
			if word.Len() > 0 {
				self.words = append(self.words, word.String())
				word.Reset()
			}
		} else {
			log.Printf("unknown: ch = %q", ch)
		}
	}
}

func (self *Reader) Read(filename string) error {
	var err error
	var data []byte
	data, err = os.ReadFile(filename)
	if err != nil {
		return err
	}

	self.data = string(data[:])

	self.parseData()

//	log.Printf("data = %+v", self.data)
//	log.Printf("words = %+v", self.words)

	return nil
}
