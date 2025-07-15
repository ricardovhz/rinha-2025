package main

import (
	"database/sql"

	"github.com/gofiber/fiber/v3"
	_ "github.com/marcboeker/go-duckdb"
	"github.com/ricardovhz/rinha-2025/internal"
)

func main() {

	db, err := sql.Open("duckdb", "./payments.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.Exec(`CREATE TABLE IF NOT EXISTS payments (
	backend INTEGER NOT NULL,
	requested_at TIMESTAMP NOT NULL,
	amount INTEGER NOT NULL
	)`)

	app := fiber.New()
	app.Post("/payments", func(c fiber.Ctx) error {
		var d internal.SavePaymentRequest
		err := c.Bind().JSON(&d)
		if err != nil {
			return err
		}

		stmt, err := db.Prepare(`INSERT INTO payments (backend, requested_at, amount) VALUES (?,?,?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(d.Backend, d.RequestedAt, d.Amount)
		if err != nil {
			return err
		}

		return nil
	})

	app.Get("/payments-summary", func(c fiber.Ctx) error {
		qfrom := c.Query("from")
		qto := c.Query("to")

		from := internal.ParseDateTime(qfrom)
		to := internal.ParseDateTime(qto)

		rows, err := db.Query(`SELECT backend, COUNT(*), SUM(amount) as total_amount FROM payments WHERE requested_at BETWEEN ? AND ? GROUP BY backend`, from, to)
		if err != nil {
			return err
		}
		defer rows.Close()

		summary := make(map[int]*internal.BackendSummary)
		summary[0] = &internal.BackendSummary{}
		summary[1] = &internal.BackendSummary{}
		for rows.Next() {
			var backend int
			var count int64
			var totalAmount int64
			if err := rows.Scan(&backend, &count, &totalAmount); err != nil {
				return err
			}
			summary[backend] = &internal.BackendSummary{
				TotalRequests: count,
				TotalAmount:   totalAmount,
			}
		}

		summaryResponse := &internal.SummaryModel{
			DefaultBackend:  summary[0],
			FallbackBackend: summary[1],
		}
		c.JSON(summaryResponse)
		c.Status(200)
		return nil
	})

	app.Listen(":3000")

}
