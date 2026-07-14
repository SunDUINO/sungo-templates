# Opis działania kodu — szablon `ginrest` (Gin REST API)

## Cel projektu

Szablon startowy REST API napisany w Go z użyciem frameworka **Gin**, część **SunGo Project Manager**. Zawiera gotową strukturę serwera HTTP z podstawowym CRUD-em, middleware'ami i mechanizmem graceful shutdown.

---

## 1. Struktura danych — `Item`

```go
type Item struct {
    ID    int
    Name  string
    Value string
}
```

Prosty model reprezentujący pojedynczy zasób. Dane trzymane są **w pamięci** (`items []Item`), bez bazy danych — to placeholder do podpięcia własnej warstwy DB.

- `nextID` — licznik służący do nadawania kolejnych ID nowym elementom.
- `mu sync.RWMutex` — **(dodane w poprawce)** chroni `items`/`nextID` przed równoczesnym dostępem z wielu goroutine obsługujących żądania HTTP.

---

## 2. Handlery (funkcje obsługujące żądania)

| Funkcja | Metoda/ścieżka | Działanie |
|---|---|---|
| `healthHandler` | `GET /health` | Zwraca status serwera, nazwę projektu i aktualny czas UTC — endpoint do health-checków (monitoring, Kubernetes). |
| `listItems` | `GET /api/v1/items` | Zwraca całą listę elementów wraz z licznikiem (`count`). Odczyt pod `mu.RLock()`. |
| `getItem` | `GET /api/v1/items/:id` | Zwraca pojedynczy element po ID. |
| `createItem` | `POST /api/v1/items` | Waliduje JSON wejściowy (`name`, `value` wymagane), tworzy nowy `Item`, dodaje do listy. Zapis pod `mu.Lock()`. |
| `deleteItem` | `DELETE /api/v1/items/:id` | Usuwa element o podanym ID. Zapis pod `mu.Lock()`. |

### Poprawki wprowadzone w kodzie

#### 🐛 Bug #1 — błędne porównanie ID

**Oryginał:**
```go
if id == string(rune('0'+item.ID)) {
```
Ten kod porównywał string z URL z pojedynczym znakiem powstałym z konwersji cyfry na `rune`. Działał przypadkiem tylko dla ID 0–9; dla ID ≥ 10 dawał błędny/nieprzewidywalny wynik i element nigdy nie zostałby znaleziony ani usunięty.

**Poprawka:**
```go
import "strconv"
...
if id == strconv.Itoa(item.ID) {
```
Teraz ID elementu jest poprawnie konwertowane na string i porównywane z parametrem z URL — działa dla dowolnej liczby cyfr. Zastosowano w `getItem` i `deleteItem`.

#### 🐛 Bug #2 — brak synchronizacji dostępu (race condition)

**Problem:** Gin domyślnie obsługuje żądania równolegle (każde w osobnej goroutine). Globalne zmienne `items` i `nextID` były modyfikowane bez żadnej blokady — przy jednoczesnych żądaniach `POST`/`DELETE` mogło dojść do:
- nadpisania się zmian (utrata danych),
- paniki przy jednoczesnym `append` + iteracji po slice,
- niedeterministycznych wartości `nextID`.

**Poprawka:** dodano `sync.RWMutex`:
- `mu.RLock()` / `mu.RUnlock()` przy odczycie (`listItems`, `getItem`),
- `mu.Lock()` / `mu.Unlock()` przy zapisie (`createItem`, `deleteItem`).

To gwarantuje, że tylko jedna goroutine na raz modyfikuje store, a odczyty mogą się odbywać równolegle względem siebie (ale nie względem zapisu).

---

## 3. Router — `setupRouter()`

Buduje instancję Gin (`gin.New()` — bez domyślnych middleware'ów) i ręcznie dokłada:

1. **`gin.Logger()`** — loguje każde żądanie HTTP (metoda, ścieżka, status, czas trwania).
2. **`gin.Recovery()`** — łapie panic-e w handlerach, zamieniając je na odpowiedź `500` zamiast crashowania całego serwera.
3. **CORS** (`gin-contrib/cors`) — pozwala na żądania z dowolnego originu (`AllowOrigins: []string{"*"}`), dozwolone metody: `GET, POST, PUT, DELETE, OPTIONS`. `AllowCredentials: false` jest spójne z `*` (przeglądarki i tak blokują `*` razem z `credentials: true`).

Trasy:
- `GET /health` — poza grupą API.
- `/api/v1/...` — grupa wersjonowanego API z 4 endpointami CRUD dla `items`.

---

## 4. Funkcja `main()` — uruchomienie i zamknięcie serwera

### Start serwera

```go
port := os.Getenv("PORT")
if port == "" { port = "8080" }
```
Port konfigurowalny przez zmienną środowiskową, domyślnie `8080`.

Serwer tworzony jest ręcznie (`http.Server`) zamiast przez `router.Run()`, co pozwala ustawić:
- `ReadTimeout`, `WriteTimeout` — 10 s (ochrona przed powolnymi klientami, np. atakiem slowloris).
- `IdleTimeout` — 60 s dla połączeń keep-alive.

Serwer nasłuchuje w **osobnej goroutine**, żeby główny wątek mógł czekać na sygnał zamknięcia.

### Graceful shutdown

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
```
Program blokuje się na kanale `quit` do momentu otrzymania sygnału `SIGINT` (Ctrl+C) lub `SIGTERM` (np. `docker stop`, `kill`).

Po sygnale:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
srv.Shutdown(ctx)
```
Serwer ma **5 sekund** na dokończenie aktualnie obsługiwanych żądań przed wymuszonym zamknięciem — standardowy wzorzec dla usług działających w kontenerach/Kubernetes.

---

## 5. Przepływ żądania

```
Klient → CORS middleware → Logger → Recovery → Router (dopasowanie trasy)
       → Handler (np. createItem) → walidacja JSON → operacja na []Item (pod mutexem)
       → JSON response
```

---

## 6. Co jeszcze warto rozważyć w realnym projekcie

- **Persystencja**: `items` żyje tylko w pamięci procesu — restart serwera = utrata danych. Docelowo podpiąć bazę danych (Postgres/SQLite/etc.).
- **Walidacja ID**: obecnie brak sprawdzenia, czy `:id` w ogóle jest liczbą — `strconv.Itoa` po stronie porównania działa poprawnie, ale warto rozważyć też `strconv.Atoi(id)` na wejściu i zwracanie `400 Bad Request` dla nieliczbowych ID.
- **CORS `*`**: OK do dev/testów; w produkcji warto ograniczyć do konkretnych originów.
- **Brak `PUT`/`PATCH`**: obecnie nie ma endpointu do aktualizacji istniejącego elementu.
- **Paginacja**: `listItems` zwraca całą listę na raz — przy większej liczbie rekordów warto dodać `limit`/`offset` lub `page`.
