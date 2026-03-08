package main

import (
	"encoding/csv"
	"flag"
	"fmt"
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
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close CSV file: %v", err)
		}
	}()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Failed to read CSV: %v", err)
		return
	}

	problems := make([]problem, len(records))
	for i, line := range records {
		problems[i] = problem{question: line[0], correct: line[1]}
	}

	for i := range problems {
		fmt.Printf("%s = ", problems[i].question)
		if _, err := fmt.Scanln(&problems[i].answer); err != nil {
			log.Printf("Failed to read answer: %v", err)
		}
	}

	problemsNum := len(problems)
	correctNum := 0
	for _, p := range problems {
		if p.correct == p.answer {
			correctNum++
		}
	}

	percentage := int(float64(correctNum) / float64(problemsNum) * 100)

	fmt.Printf("You scored %d out of %d (%d%%).\n", correctNum, problemsNum, percentage)
}
