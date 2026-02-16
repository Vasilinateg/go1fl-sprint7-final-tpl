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

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	tests := []struct {
		name  string
		count int
		want  int
	}{
		{
			name:  "count=0 (пустой ответ)",
			count: 0,
			want:  0,
		},
		{
			name:  "count=1 (одно кафе)",
			count: 1,
			want:  1,
		},
		{
			name:  "count=2 (два кафе)",
			count: 2,
			want:  2,
		},
		{
			name:  "count=100 (больше чем есть)",
			count: 100,
			want:  len(cafeList["moscow"]),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/cafe?city=moscow&count=" + strconv.Itoa(tt.count)
			req := httptest.NewRequest(http.MethodGet, url, nil)

			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)

			body := strings.TrimSpace(w.Body.String())

			var cafes []string
			if body != "" {
				cafes = strings.Split(body, ",")
			}

			assert.Equal(t, tt.want, len(cafes))
		})
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	tests := []struct {
		name      string
		search    string
		wantCount int
	}{
		{
			name:      "поиск 'фасоль' (нет результатов)",
			search:    "фасоль",
			wantCount: 0,
		},
		{
			name:      "поиск 'кофе' (2 кафе)",
			search:    "кофе",
			wantCount: 2,
		},
		{
			name:      "поиск 'вилка' (1 кафе)",
			search:    "вилка",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/cafe?city=moscow&search=" + tt.search
			req := httptest.NewRequest(http.MethodGet, url, nil)

			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)

			body := strings.TrimSpace(w.Body.String())

			if body == "" {
				assert.Equal(t, 0, tt.wantCount)
				return
			}

			cafes := strings.Split(body, ",")

			assert.Equal(t, tt.wantCount, len(cafes))

			searchLower := strings.ToLower(tt.search)
			for _, cafe := range cafes {
				cafeLower := strings.ToLower(cafe)
				assert.Contains(t, cafeLower, searchLower,
					"Кафе '%s' не содержит '%s'", cafe, tt.search)
			}
		})
	}
}
