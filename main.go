package main

import (
    "database/sql"
    "fmt"
	"log"
	"os"
	"strings"
	"strconv"
	"github.com/lib/pq"
)

const (
    host     = "localhost"
    port     = 5432
    user     = "postgres"
    password = "password"
    dbname   = "reDashha"
)

func main() {
    data_base := fmt.Sprintf("host=%s port=%d user=%s "+
        "password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname)
    db, err := sql.Open("postgres", data_base)
    if err != nil {
		log.Fatal(err)
	}
    defer db.Close()
  
    err = db.Ping()
    if err != nil {
		log.Fatal(err)
	}
	
	args:= os.Args[1:]
	orders_str := strings.Split(args[0], ",")
	
	orders_int := make([]int, len(orders_str))
	for i, v := range orders_str {
        num, err := strconv.Atoi(v)
		if err != nil {
			log.Fatal(err)
		}
        orders_int[i] = num
    }

	query := `
	SELECT rack.name AS shelf, 
		CONCAT(item.name, ' (id=', item.id, ')') AS item_with_id, 
		CONCAT('заказ ', orders.order_number, ', ', orders.amount, ' шт') AS order_with_amount,  
		COALESCE((SELECT string_agg(rack.name, ',')
            FROM rack
                JOIN item_in_rack ON rack_id = rack.id
            WHERE
                main_rack = FALSE
                AND item_in_rack.item_id = item.id
        ), '-' 
    ) AS add_shelf 
	FROM rack 
	JOIN item_in_rack ON rack_id = rack.id 
	JOIN item ON item_in_rack.item_id = item.id 
	JOIN orders ON orders.item_id = item.id 
	WHERE main_rack = TRUE AND order_number = ANY ($1::INT[])
	ORDER BY shelf, orders.id`

	rows, err := db.Query(query, pq.Array(orders_int))

	defer rows.Close()

	fmt.Printf("=+=+=+=\nСтраница сборки заказов %s\n\n", args[0])
	current_unit := ""
	for rows.Next() {
		var shelf, item_with_id, order_with_amount, add_shelf string
		err := rows.Scan(&shelf, &item_with_id, &order_with_amount, &add_shelf)
		if err != nil {
			log.Fatal(err)
		}
        if current_unit != shelf {
            fmt.Printf("===Стеллаж %s\n", shelf)
            current_unit = shelf
        }
        fmt.Printf("%s\n%s\n", item_with_id, order_with_amount)
		if add_shelf != "-" {
			fmt.Printf("дополнительные стеллажи: %s\n\n", add_shelf)
		} else {
			fmt.Printf("\n")
		}
    }
}