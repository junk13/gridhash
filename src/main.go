/**
 * Created by adamcaudill on 12/27/13.
 */
package main

import (
	"fmt"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"
	"math/rand"
)

var grid [][][]byte
var extra_data[][]byte

func main() {
	//create main hash
	start := time.Now()
	initial_hash := initial_pwd_hash("test")
	hash := gridhash(initial_hash, 48, 1000, 100, []byte("testtest"), 10000)
	elapsed := time.Since(start)

	fmt.Println("Initial: ", hex.EncodeToString(initial_hash))
	fmt.Println("Final: ", hex.EncodeToString(hash))
	fmt.Printf("Hash took %s\n", elapsed)
}

func gridhash(password []byte, grid_size int, hmac_iter int, hash_iter int, salt []byte, extra_bytes int) []byte {
	//setup grid to hold hashes
	grid = make([][][]byte, grid_size)
	for i := range grid {
		grid[i] = make([][]byte, grid_size)
	}

	//create extra data
	start := time.Now()
	extra_data = make([][]byte, grid_size)
	for i := 0; i < grid_size; i++ {
		//todo: better seed
		extra_data[i] = rand_bytes(extra_bytes, int64(i))
	}
	elapsed := time.Since(start)
	fmt.Printf("Extra data generation took %s\n", elapsed)

	//time to crunch...
	for i := 0; i < grid_size; i++ {
		round(i, password, hmac_iter, hash_iter, salt)
	}

	return grid[grid_size-1][grid_size-1]
}

func round(index int, password []byte, hmac_iter int, hash_iter int, salt []byte) {
	//fmt.Println("Index: ", index)
	if (index == 0) {
    //special case to bootstrap the process for later

		//set 0:0
		grid[index][index] = kdf(password, salt, password, hmac_iter)

		//set 0:1
		grid[index][index+1] = hash(append(grid[index][index], salt...), hash_iter)
	} else {
    //top down
		for y := 0; y <= index-1; y++ {
			process_cell(index, y, hmac_iter, hash_iter, salt, password)
		}

		//left to right
		for x := 0; x <= index; x++ {
			process_cell(x, index, hmac_iter, hash_iter, salt, password)
		}
	}
}

func process_cell(row int, column int, hmac_iter int, hash_iter int, salt []byte, password []byte) {
  //make sure this isn't a cell that we bootstrapped
	if (grid[row][column] == nil) {
		var value []byte
		var ret []byte

		//first, the pass going bottom up
		for idx := row - 1; idx >= 0; idx-- {
			value = append(value, grid[idx][column]...)
		}

		//next, go right to left
		for idx := column - 1; idx >= 0; idx-- {
			value = append(value, grid[row][idx]...)
		}

		//add the extra data
		value = append(value, extra_data[column]...)

		//is this a key cell?
		if (row == column) {
			ret = kdf(value, salt, password, hmac_iter)
		} else {
			ret = hash(append(value, salt...), hash_iter)
		}

		grid[row][column] = ret
		//fmt.Println("Cell: ", row, column, hex.EncodeToString(grid[row][column]))
	}
}

func initial_pwd_hash(password string) []byte {
	pwd := []byte(password)
	mac := hmac.New(sha256.New, hash_sum(pwd))
	mac.Write(pwd)
	return mac.Sum(nil)
}

func kdf(value []byte, salt []byte, password []byte, iterations int) []byte {
  for i := 0; i < iterations; i++ {
		mac := hmac.New(sha256.New, password)
		mac.Write(append(value, salt...))
		value = mac.Sum(nil)
	}

	return value
}

func hash(value []byte, iterations int) []byte {
	for i := 0; i < iterations; i++ {
		value = hash_sum(value)
	}

	return value
}

func hash_sum(data []byte) []byte {
	h := sha256.Sum256(data)
	s := string(h[:])
	return []byte(s)
}

func rand_bytes(length int, seed int64) []byte {
	//todo: this is slow, and Go doesn't make it easy
	r := rand.New(rand.NewSource(seed))

	var list []byte
	for i := 0; i < length; i++ {
		buff := byte(r.Intn(255))
		list = append(list, buff)
	}

	return list
}
