// ============ PHẦN 1: GENERATE .BIN (Backend Go) ============
// File: cmd/generate_calendar_bin.go

package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
	"remiaq/internal/services"
	"time"
)

const (
	OutputFile = "calendar_data.bin"
	StartYear  = 2025
	EndYear    = 2035
)

type BinaryLunarDate struct {
	Year   int16 // 2 bytes
	Month  uint8 // 1 byte
	Day    uint8 // 1 byte
	IsLeap uint8 // 1 byte (0 hoặc 1)
}

func main() {
	lc := services.NewLunarCalendar()
	start := time.Date(StartYear, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(EndYear, 12, 31, 23, 59, 59, 0, time.UTC)

	// Xóa file cũ
	os.Remove(OutputFile)

	file, err := os.Create(OutputFile)
	if err != nil {
		log.Fatal("Create file failed:", err)
	}
	defer file.Close()

	var buffer bytes.Buffer
	count := 0

	// Header: số lượng entries (4 bytes)
	// Dữ liệu từ StartYear-01-01 đến EndYear-12-31
	numDays := int(end.Sub(start).Hours()/24) + 1

	// Ghi header
	err = binary.Write(&buffer, binary.LittleEndian, int32(numDays))
	if err != nil {
		log.Fatal("Write header failed:", err)
	}

	// Ghi từng ngày
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		vnTime := d.In(time.FixedZone("VN", 7*3600))
		lunar := lc.SolarToLunar(vnTime)

		isLeap := uint8(0)
		if lunar.IsLeap {
			isLeap = 1
		}

		// Tính ngày thứ mấy (0 = Jan 1, 2024)
		dayIndex := int32(d.Sub(start).Hours() / 24)

		// Ghi định dạng: [dayIndex(4)] [year(2)] [month(1)] [day(1)] [isLeap(1)]
		err = binary.Write(&buffer, binary.LittleEndian, dayIndex)
		if err != nil {
			log.Fatal("Write dayIndex failed:", err)
		}

		err = binary.Write(&buffer, binary.LittleEndian, int16(lunar.Year))
		if err != nil {
			log.Fatal("Write year failed:", err)
		}

		err = binary.Write(&buffer, binary.LittleEndian, uint8(lunar.Month))
		if err != nil {
			log.Fatal("Write month failed:", err)
		}

		err = binary.Write(&buffer, binary.LittleEndian, uint8(lunar.Day))
		if err != nil {
			log.Fatal("Write day failed:", err)
		}

		err = binary.Write(&buffer, binary.LittleEndian, isLeap)
		if err != nil {
			log.Fatal("Write isLeap failed:", err)
		}

		count++
	}

	// Ghi vào file
	file.Write(buffer.Bytes())

	log.Printf("✅ Binary file generated: %s", OutputFile)
	log.Printf("   Entries: %d", count)
	log.Printf("   File size: %d bytes", buffer.Len())
	log.Printf("   Period: %s to %s", start.Format("2006-01-02"), end.Format("2006-01-02"))
}

// ============ PHẦN 2: READ .BIN (Android Kotlin) ============
// File: CalendarBinaryReader.kt

/*
package quangan.oreminde.data

import android.content.Context
import java.io.DataInputStream
import java.time.LocalDate

data class LunarDate(
    val year: Int,
    val month: Int,
    val day: Int,
    val isLeapMonth: Boolean = false
)

class CalendarBinaryReader(private val context: Context) {
    private lateinit var lunarCache: Map<LocalDate, LunarDate>
    private val baseDate = LocalDate.of(2024, 1, 1) // Ngày bắt đầu trong file .bin

    init {
        loadBinaryData()
    }

    private fun loadBinaryData() {
        val cache = mutableMapOf<LocalDate, LunarDate>()

        try {
            val inputStream = context.assets.open("calendar_data.bin")
            val dataInput = DataInputStream(inputStream)

            // Đọc header: số lượng entries
            val numEntries = dataInput.readInt()

            // Đọc từng entry
            repeat(numEntries) {
                val dayIndex = dataInput.readInt()
                val year = dataInput.readShort().toInt()
                val month = dataInput.readUnsignedByte()
                val day = dataInput.readUnsignedByte()
                val isLeap = dataInput.readUnsignedByte() == 1

                val solarDate = baseDate.plusDays(dayIndex.toLong())
                val lunar = LunarDate(year, month, day, isLeap)

                cache[solarDate] = lunar
            }

            lunarCache = cache
            dataInput.close()

        } catch (e: Exception) {
            e.printStackTrace()
        }
    }

    fun getLunarDate(solarDate: LocalDate): LunarDate? {
        return lunarCache[solarDate]
    }

    fun getCacheSize(): Int = lunarCache.size
}

// Sử dụng:
// val reader = CalendarBinaryReader(context)
// val lunar = reader.getLunarDate(LocalDate.now())
*/
