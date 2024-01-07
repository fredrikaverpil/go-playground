package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

type problem struct {
	question string
	correct  string
	answer   string
}

func main() {
	csvFile := flag.String("csv", "problems.csv", "a csv file in the format of 'question,answer'")
	flag.Parse()

	file, err := os.Open(*csvFile)
	if err != nil {
		fmt.Printf("Failed to open the CSV file: %s\n", *csvFile)
		os.Exit(1)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	var problems []problem

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Print(err)
			os.Exit(1)
		}

		problems = append(problems, problem{question: line[0], correct: line[1]})
	}

	for i, p := range problems {
		fmt.Printf("%s = ", p.question)
		fmt.Scanln(&p.answer)
		problems[i].answer = p.answer
	}

	problems_num := len(problems)
	correct_num := 0
	for _, p := range problems {
		if p.correct == p.answer {
			correct_num++
		}
	}

	percentage := int(float64(correct_num) / float64(problems_num) * 100)

	fmt.Printf("You scored %d out of %d (%d%%).\n", correct_num, problems_num, percentage)
}
