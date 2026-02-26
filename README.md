# Spam Filter

A simple Naive Bayes spam filter built in Go.

## What it does

- Trains on email data from the Enron dataset (folders 1-5)
- Uses bag-of-words approach with Laplace smoothing
- Classifies emails as spam or ham
- Outputs accuracy metrics, confusion matrix, and ASCII bar charts

## How to run

```bash
go build -o spam-filter .
./spam-filter
```

## Files

- `main.go` - Core classification logic
- `report.go` - Results and visualization

## Data

Uses Enron email dataset. Training on enron1-5, testing on enron6.

## References

- https://en.wikipedia.org/wiki/Naive_Bayes_spam_filtering
- https://en.wikipedia.org/wiki/Bag-of-words_model
- http://www2.aueb.gr/users/ion/data/enron-spam/
