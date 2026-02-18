package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// --- ОРИГИНАЛЬНАЯ ЛОГИКА ИГРЫ (из вашего файла) ---

type Character interface {
	hit_target(target Character, hit_point string)
	block_attack(hit_point string) bool
	get_hp() int
	get_name() string
}

type item struct {
	type_item string
	name      string
	attack    int
	defence   int
	plus_hp   int
}

type player struct {
	name      string
	hp        int
	strength  int
	hit       string
	block     string
	inventory []item
	equipment []item
}

type enemy struct {
	name     string
	hp       int
	strength int
	hit      string
	block    string
	trophy   item
}

func (p *player) get_name() string { return p.name }
func (p *player) get_hp() int      { return p.hp }
func (e *enemy) get_name() string  { return e.name }
func (e *enemy) get_hp() int       { return e.hp }

func (p *player) hit_target(target Character, hit_point string) {
	damage := p.strength
	for _, it := range p.equipment {
		if it.type_item == "оружие" {
			damage += it.attack
		}
	}
	switch hit_point {
	case "уши":
		damage = 5
	case "глаза":
		damage = 10
	}
	if !target.block_attack(hit_point) {
		if t, ok := target.(*enemy); ok {
			t.hp -= damage
		}
		if t, ok := target.(*player); ok {
			t.hp -= damage
		}
	}
}

func (p *player) block_attack(hit_point string) bool {
	return p.block == hit_point
}

func (e *enemy) hit_target(target Character, hit_point string) {
	if !target.block_attack(hit_point) {
		if t, ok := target.(*player); ok {
			t.hp -= e.strength
		}
	}
}

func (e *enemy) block_attack(hit_point string) bool {
	return e.block == hit_point
}

// --- СЕТЕВАЯ ЧАСТЬ (СЕРВЕР) ---

var (
	game_log            []string
	log_mutex           sync.Mutex
	client_move         chan string = make(chan string, 1)
	is_client_connected bool
)

func game_handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		log_mutex.Lock()
		fmt.Fprint(w, strings.Join(game_log, "\n"))
		log_mutex.Unlock()
	} else if r.Method == http.MethodPost {
		body, _ := io.ReadAll(r.Body)
		client_move <- string(body)
		is_client_connected = true
		fmt.Fprint(w, "OK")
	}
}

func add_log(msg string) {
	log_mutex.Lock()
	game_log = append(game_log, msg)
	fmt.Println(msg)
	log_mutex.Unlock()
}

func play_network_pvp_server() {
	p1 := player{name: "Сервер (P1)", hp: 100, strength: 10}
	p2 := player{name: "Клиент (P2)", hp: 100, strength: 10}

	add_log("Ожидание подключения клиента...")
	for !is_client_connected {
		time.Sleep(1 * time.Second)
	}
	add_log("Клиент подключен! Начинаем бой.")

	scanner := bufio.NewScanner(os.Stdin)
	for p1.hp > 0 && p2.hp > 0 {
		add_log(fmt.Sprintf("\n--- РАУНД: %s (%d HP) vs %s (%d HP) ---", p1.name, p1.hp, p2.name, p2.hp))

		// Ход сервера
		fmt.Println("Ваш ход (атака: голова/тело/ноги, блок: голова/тело/ноги):")
		scanner.Scan()
		input := strings.Split(scanner.Text(), " ")
		p1.hit, p1.block = input[0], input[1]

		add_log("Ожидание хода клиента...")
		c_input := strings.Split(<-client_move, " ")
		p2.hit, p2.block = c_input[0], c_input[1]

		p1.hit_target(&p2, p1.hit)
		p2.hit_target(&p1, p2.hit)

		add_log(fmt.Sprintf("Результат: %s ударил в %s, %s ударил в %s", p1.name, p1.hit, p2.name, p2.hit))
	}
	if p1.hp <= 0 {
		add_log("Клиент победил!")
	} else {
		add_log("Сервер победил!")
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("1. Сюжет\n2. PvP (Hotseat)\n3. Сетевой PvP (Сервер)")
	var choice int
	fmt.Scan(&choice)

	if choice == 3 {
		http.HandleFunc("/", game_handler)
		go http.ListenAndServe(":8080", nil) // Запуск HTTP-сервера
		play_network_pvp_server()
	} else if choice == 2 {
		// Тут ваша оригинальная функция play_pvp()
		fmt.Println("Запуск локального PvP...")
	} else {
		// Тут ваша оригинальная функция play_story()
		fmt.Println("Запуск сюжета...")
	}
}
