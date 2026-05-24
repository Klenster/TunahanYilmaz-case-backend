# PassportX — Backend

Go + Gin ile yazılmış PassportX uygulamasının API'si.

---

## Teknoloji Seçimleri ve Gerekçeleri

### Dil - Go

Python veya Node.js'e kıyasla derlenen binary üretir, runtime kurulumu gerekmez. Statik tip sistemi hataları derleme zamanında yakalar, Goroutine ile yüksek eşzamanlılık sağlar,düşük bellek kullanılır.

### Framework - Gin

Go'nun en yaygın web framework'ü. Express.js'e kıyasla çok daha düşük latency, middleware zinciri, route gruplama ve JSON binding built-in gelir; fazladan kütüphanelere gerek kalmaz.

### Orm - Gorm

SQLAlchemy veya Prisma'ya alternatiftir. AutoMigrate ile migration dosyası yazmadan şema yönetimi, Preload ile ilişkisel veri desteği sağlar.

### Veritabanı - SQLite

PostgreSQL veya MySQL'e kıyasla daha kolay kurulum, tek dosya. Bu ölçekteki bir proje için en uygun seçim; ileride sadece driver değiştirilerek PostgreSQL'e geçilebilir.

### Auth - JWT

Session tabanlı auth'a kıyasla sunucuda state tutmaz.Token içerisinde rol bilgisi taşıdığı için her istekte kullanıcı sorgusu yapmaya gerek kalmaz, performans artar.

### Şifreleme - bcrypt

MD5/SHA gibi hızlı algoritmaların aksine kasıtlı yavaş çalıştırılır, brute-force olasılığını azaltır.Salt otomatik üretilir,rainbow table saldırılarını önleyebilir.

### UUID - google/uuid

Integer ID'ye kıyasla tahmin edilemez. saldırgan /products/1, /products/2 diyerek kayıtları enumerate edemez.

## Kurulum

### Gereksinimler

- Go 1.22+

### Adımlar

```bash
# 1.klasöre geçiş
cd TunahanYilmaz-case-backend

# 2. Bağımlılıkları indir
go mod tidy

# 3. Ortam değişkenlerini ayarla
cp .env.example .env
# .env dosyasını aç, APP_SECRET_KEY'i güçlü bir değerle doldur

# 4. Uygulamayı çalıştır
go run ./cmd/server
```

API **http://localhost:8000** adresinde çalışacaktır.

---

## Varsayılan Admin Kimlik Bilgileri

### Üretim Yaklaşımı

Uygulama ilk kez çalıştırıldığında sistemde admin kullanıcısı yoksa otomatik olarak oluşturulur (`internal/db/db.go` — `SeedAdmin` fonksiyonu).

Şifre üretim süreci:

- `.env` dosyasında `ADMIN_PASSWORD` tanımlanmışsa o kullanılır
- Tanımlanmamışsa Go'nun `crypto/rand` paketi ile 20 karakterlik rastgele şifre üretilir
- Şifre koda gömülmez, README'ye yazılmaz, repoya commit edilmez
- **Yalnızca ilk başlatmada stdout'a bir kez yazdırılır:**

```
============================================================
DEFAULT ADMIN CREDENTIALS (save these now!)
  Email:    admin@passportx.com
  Password: A1!xK9mR2vLqN5pT8wYz
These will NOT be shown again.
============================================================
```

- Şifre bcrypt ile hashlenerek saklanır, veritabanında düz metin yoktur
- Production'da stdout bir log dosyasına yönlendirilerek güvenli şekilde saklanabilir
- `.env`'de `ADMIN_PASSWORD` tanımlayarak önceden belirlenmiş şifre kullanılabilir

---

## Proje Yapısı

```
backend/
├── cmd/
│   └── server/
│       └── main.go          # Uygulama girişi, router ve middleware kurulumu
├── internal/
│   ├── config/
│   │   └── config.go        # .env okuma, uygulama ayarları
│   ├── db/
│   │   └── db.go            # GORM bağlantısı, AutoMigrate, admin seeder
│   ├── handlers/
│   │   ├── auth.go          # Register, login, me endpoint'leri
│   │   ├── products.go      # Ürün CRUD, stats, categories
│   │   └── users.go         # Kullanıcı yönetimi, profil, şifre değiştirme
│   ├── middleware/
│   │   └── auth.go          # JWT doğrulama, admin rol kontrolü
│   └── models/
│       └── models.go        # GORM modelleri: User, Product, Material
└── pkg/
    └── utils/
        └── jwt.go           # JWT üretme ve doğrulama
```

---

## API Endpoint'leri

| Method | Path                       | Yetki          | Açıklama                       |
| ------ | -------------------------- | -------------- | ------------------------------ |
| POST   | `/api/auth/register`       | Herkese açık   | Kayıt (auditor rolü)           |
| POST   | `/api/auth/login`          | Herkese açık   | Giriş, JWT döner               |
| GET    | `/api/auth/me`             | Giriş yapılmış | Mevcut kullanıcı               |
| GET    | `/api/products/`           | Giriş yapılmış | Ürün listesi (filtrelenebilir) |
| POST   | `/api/products/`           | Admin          | Ürün oluştur                   |
| GET    | `/api/products/stats`      | Giriş yapılmış | Dashboard istatistikleri       |
| GET    | `/api/products/categories` | Giriş yapılmış | Kategoriler                    |
| GET    | `/api/products/:id`        | Giriş yapılmış | Ürün detayı                    |
| PUT    | `/api/products/:id`        | Admin          | Ürün güncelle                  |
| DELETE | `/api/products/:id`        | Admin          | Ürün sil                       |
| GET    | `/api/users/`              | Admin          | Kullanıcı listesi              |
| PATCH  | `/api/users/:id/role`      | Admin          | Rol değiştir                   |
| DELETE | `/api/users/:id`           | Admin          | Kullanıcı sil                  |
| PATCH  | `/api/users/me/profile`    | Giriş yapılmış | Profil güncelle                |
| PATCH  | `/api/users/me/password`   | Giriş yapılmış | Şifre değiştir                 |

---

## Rol Bazlı Yetki Kontrolü

Yetki kontrolü **backend'de** uygulanır, auditorlar frontend'den izinsiz erişim gerçekleştiremez:

- `AuthMiddleware` — JWT token'ı doğrular, kullanıcı ID ve rolünü Gin context'ine ekler
- `AdminMiddleware` — Rol kontrolü yapar, auditor ise `403 Forbidden` döner

Auditor bir kullanıcı doğrudan `POST /api/products/` isteği atsa bile `403` alır.

---

## Güvenlik Notları

- Şifreler bcrypt ile hashlenir, veritabanında asla düz metin saklanmaz
- JWT token'lar 24 saat geçerlidir (`.env`'den ayarlanabilir)
- `.env` dosyası `.gitignore`'da, repoya commit edilmez
- Kullanıcı bulunamadı ile yanlış şifre aynı hata mesajını döner (kullanıcı enumeration önleme)
- Materyal yüzdesi validasyonu backend'de yapılır

---

## Varsayımlar

- SQLite tek sunucu için yeterlidir; çok sunuculu deployment için PostgreSQL tercih edilmeli
- Materyal kompozisyonu toplamının %100 olması zorunludur (±0.01 tolerans)
- Register endpoint'inden oluşturulan tüm kullanıcılar varsayılan olarak `auditor` rolüyle açılır
