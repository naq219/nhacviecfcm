// Command gen_lunar_table sinh bang chuyen doi Am/Duong tu thuat toan goc
// (internal/services/lunar_calendar.go) va ghi ra file TypeScript de Nuxt dung.
//
// Cach chay (tu thu muc goc project Go):
//   go run ./cmd/gen_lunar_table -from 2026-08-01 -to 2029-08-31 -out "<duong dan>\lunar-table.ts"
//
// Luu y: BANG NAY CO HAN (chi phu khoang -from..-to). Het han thi chay lai lenh tren.
// Header file sinh ra chi dung ASCII de tranh loi encoding tren Windows.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"remiaq/internal/services"
)

const (
	vnTimeZone = 7.0
	layoutIn   = "2006-01-02"
)

// [day, month, year, isLeap]
type lunarTuple [4]int

func main() {
	from := flag.String("from", "2026-08-01", "ngay bat dau (YYYY-MM-DD, duong lich)")
	to := flag.String("to", "2029-08-31", "ngay ket thuc (YYYY-MM-DD, duong lich)")
	out := flag.String("out", "", "duong dan file .ts dich")
	flag.Parse()

	if *out == "" {
		fmt.Fprintln(os.Stderr, "thieu -out")
		os.Exit(1)
	}

	start, err := time.Parse(layoutIn, *from)
	if err != nil {
		panic(err)
	}
	end, err := time.Parse(layoutIn, *to)
	if err != nil {
		panic(err)
	}
	if end.Before(start) {
		fmt.Fprintln(os.Stderr, "-to phai sau -from")
		os.Exit(1)
	}

	// ---- Pass 1: quet mot khoang RONG HON de phat hien thang nhuan ----
	// ConvertLunar2Solar khong tu choi thang nhuan o nam KHONG nhuan (tra ve ngay sai),
	// nen ta chi duoc phep sinh khoa "-true" cho (nam, thang) that su la thang nhuan.
	detectStart := start.AddDate(-2, 0, 0)
	detectEnd := end.AddDate(2, 0, 0)
	leapMonthByYear := map[int]int{}
	minYear, maxYear := 9999, 0

	for d := detectStart; !d.After(detectEnd); d = d.AddDate(0, 0, 1) {
		_, lm, ly, leap := services.ConvertSolar2Lunar(d.Day(), int(d.Month()), d.Year(), vnTimeZone)
		if leap != 0 {
			leapMonthByYear[ly] = lm
		}
		if ly < minYear {
			minYear = ly
		}
		if ly > maxYear {
			maxYear = ly
		}
	}

	// ---- Pass 2: duyet khoang can xuat -> ngay am ----
	solarToLunar := map[string]lunarTuple{}
	lunarToSolar := map[string]string{}

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		ld, lm, ly, leap := services.ConvertSolar2Lunar(d.Day(), int(d.Month()), d.Year(), vnTimeZone)
		key := d.Format(layoutIn)
		solarToLunar[key] = lunarTuple{ld, lm, ly, leap}

		lkey := fmt.Sprintf("%d-%d-%d", ly, lm, ld)
		if _, exists := lunarToSolar[lkey]; !exists {
			lunarToSolar[lkey] = key
		}
	}

	// ---- Pass 3: sinh chieu Am -> Duong ----
	// Bao gom ca nam am ke tiep de ham "tim ngay am thang sau" khong bi vo bang.
	for ly := minYear; ly <= maxYear+1; ly++ {
		leapMonth, hasLeap := leapMonthByYear[ly]
		for lm := 1; lm <= 12; lm++ {
			for ld := 1; ld <= 30; ld++ {
				dd, mm, yy := services.ConvertLunar2Solar(ld, lm, ly, 0, vnTimeZone)
				if dd == 0 || mm == 0 || yy == 0 {
					continue
				}
				lunarToSolar[fmt.Sprintf("%d-%d-%d", ly, lm, ld)] =
					fmt.Sprintf("%04d-%02d-%02d", yy, mm, dd)

				// Chi sinh thang nhuan neu (ly, lm) that su la thang nhuan
				if hasLeap && lm == leapMonth {
					if d2, m2, y2 := services.ConvertLunar2Solar(ld, lm, ly, 1, vnTimeZone); d2 != 0 {
						lunarToSolar[fmt.Sprintf("%d-%d-%d-true", ly, lm, ld)] =
							fmt.Sprintf("%04d-%02d-%02d", y2, m2, d2)
					}
				}
			}
		}
	}

	// ---- Ghi ra .ts (ASCII only o header) ----
	var sb strings.Builder
	sb.WriteString("// FILE DUOC SINH TU DONG - KHONG SUA TAY\n")
	sb.WriteString("// Nguon: internal/services/lunar_calendar.go (Jean Meeus / Ho Ngoc Duc), timezone +07\n")
	sb.WriteString(fmt.Sprintf("// Sinh bang: go run ./cmd/gen_lunar_table -from %s -to %s\n", *from, *to))
	sb.WriteString(fmt.Sprintf("// Phu: %d ngay duong (%s..%s); %d khoa am (nam am %d..%d)\n",
		len(solarToLunar), *from, *to, len(lunarToSolar), minYear, maxYear+1))
	if len(leapMonthByYear) > 0 {
		ys := make([]int, 0, len(leapMonthByYear))
		for y := range leapMonthByYear {
			ys = append(ys, y)
		}
		sort.Ints(ys)
		parts := make([]string, 0, len(ys))
		for _, y := range ys {
			parts = append(parts, fmt.Sprintf("%d:thang %d", y, leapMonthByYear[y]))
		}
		sb.WriteString("// Thang nhuan: " + strings.Join(parts, ", ") + "\n")
	}
	sb.WriteString("// Khoa am: \"YYYY-M-D\" hoac \"YYYY-M-D-true\" neu thang nhuan\n\n")

	sb.WriteString("export type LunarTuple = [day: number, month: number, year: number, isLeap: 0 | 1]\n\n")

	b, _ := json.Marshal(solarToLunar)
	sb.WriteString("export const SOLAR_TO_LUNAR: Record<string, LunarTuple> = ")
	sb.Write(b)
	sb.WriteString("\n\n")

	lkeys := make([]string, 0, len(lunarToSolar))
	for k := range lunarToSolar {
		lkeys = append(lkeys, k)
	}
	sort.Strings(lkeys)
	ordered := make(map[string]string, len(lunarToSolar))
	for _, k := range lkeys {
		ordered[k] = lunarToSolar[k]
	}
	b2, _ := json.Marshal(ordered)
	sb.WriteString("export const LUNAR_TO_SOLAR: Record<string, string> = ")
	sb.Write(b2)
	sb.WriteString("\n")

	if err := os.WriteFile(*out, []byte(sb.String()), 0o644); err != nil {
		panic(err)
	}

	fmt.Printf("Da ghi %s\n", *out)
	fmt.Printf("  solarToLunar : %d ngay (%s .. %s)\n", len(solarToLunar), *from, *to)
	fmt.Printf("  lunarToSolar : %d khoa (nam am %d .. %d)\n", len(lunarToSolar), minYear, maxYear+1)
	ys := make([]int, 0, len(leapMonthByYear))
	for y := range leapMonthByYear {
		ys = append(ys, y)
	}
	sort.Ints(ys)
	for _, y := range ys {
		fmt.Printf("  nam am %d nhuan thang %d\n", y, leapMonthByYear[y])
	}
}
