package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCafeCount проверяет работу параметра count
func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	tests := []struct {
		name     string
		city     string
		count    int
		expected int
	}{
		{"Count=0 Moscow", "moscow", 0, 0},
		{"Count=1 Moscow", "moscow", 1, 1},
		{"Count=2 Moscow", "moscow", 2, 2},
		{"Count=100 Moscow", "moscow", 100, len(cafeList["moscow"])},
		{"Count=0 Tula", "tula", 0, 0},
		{"Count=1 Tula", "tula", 1, 1},
		{"Count=100 Tula", "tula", 100, len(cafeList["tula"])},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/cafe?city=" + tt.city + "&count=" + strconv.Itoa(tt.count)
			req := httptest.NewRequest("GET", url, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)
			response := rr.Body.String()

			// Для count=0 ожидаем пустую строку
			if tt.count == 0 {
				assert.Empty(t, response)
				return
			}

			cafes := strings.Split(response, ",")
			assert.Equal(t, tt.expected, len(cafes))
		})
	}
}

// TestCafeSearch проверяет работу параметра search
func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	tests := []struct {
		name        string
		search      string
		expectedLen int
	}{
		{"Search=фасоль", "фасоль", 0},
		{"Search=кофе", "кофе", 2},         // "Мир кофе" и "Кофе и завтраки"
		{"Search=вилка", "вилка", 1},       // "Ложка и вилка"
		{"Search=Завтраки", "Завтраки", 1}, // "Кофе и завтраки"
		{"Search=мир", "мир", 1},           // Только "Мир кофе" (в "Кофе и завтраки" нет "мир")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/cafe?city=moscow&search=" + tt.search
			req := httptest.NewRequest("GET", url, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)
			response := rr.Body.String()

			if tt.expectedLen == 0 {
				assert.Empty(t, response)
				return
			}

			cafes := strings.Split(response, ",")
			assert.Equal(t, tt.expectedLen, len(cafes))

			lowerSearch := strings.ToLower(tt.search)
			for _, cafe := range cafes {
				lowerCafe := strings.ToLower(cafe)
				assert.True(t, strings.Contains(lowerCafe, lowerSearch),
					"Кафе '%s' не содержит подстроку '%s'", cafe, tt.search)
			}
		})
	}
}

// Существующие тесты остаются без изменений
func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}
