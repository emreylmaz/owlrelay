# 🦉 OwlRelay Roadmap

**Son Güncelleme:** 2026-02-11  
**Mevcut Versiyon:** v0.1.1

---

## Phase 1: MVP ✅ (v0.1.x) — TAMAMLANDI

| Özellik | Durum | Açıklama |
|---------|-------|----------|
| Relay Server (Go) | ✅ | HTTP + WebSocket, chi router |
| Token Auth | ✅ | SHA-256 hash, SQLite |
| Rate Limiting | ✅ | 100 req/min per token |
| Chrome Extension | ✅ | Manifest V3, TypeScript |
| Click Command | ✅ | Selector veya coordinates |
| Type Command | ✅ | Input'a text yazma |
| Scroll Command | ✅ | up/down/left/right |
| Screenshot | ✅ | PNG capture |
| DOM Snapshot | ✅ | Simplified HTML |
| Single Tab | ✅ | Bir tab'a bağlanma |
| Docker Deploy | ✅ | ~10MB image |
| GitHub Release | ✅ | v0.1.1 |

---

## Phase 2: Core Enhancements (v0.2.0) — SIRADA

**Hedef:** 1 hafta  
**Öncelik:** Yüksek

### 2.1 Multi-Tab Support
| Task | Effort | Açıklama |
|------|--------|----------|
| Extension: Multiple tab attach | 2h | Birden fazla tab'ı aynı anda track et |
| API: Tab selection | 1h | Her komutta tabId zorunlu |
| UI: Tab list improvements | 1h | Popup'ta tüm tab'ları göster/yönet |
| Tests | 1h | Multi-tab senaryoları |

### 2.2 Wait Conditions
| Task | Effort | Açıklama |
|------|--------|----------|
| waitForSelector | 1h | Element görünene kadar bekle |
| waitForText | 1h | Belirli text görünene kadar bekle |
| waitForNavigation | 1h | Sayfa yüklenene kadar bekle |
| waitForNetwork | 2h | XHR/fetch tamamlanana kadar bekle |
| Timeout handling | 0.5h | Configurable timeout |

### 2.3 Keyboard Shortcuts
| Task | Effort | Açıklama |
|------|--------|----------|
| press() command | 1h | Tek tuş: Enter, Tab, Escape |
| Modifier keys | 1h | Ctrl+A, Ctrl+C, Ctrl+V |
| Key sequences | 1h | Birden fazla tuş kombinasyonu |
| Special keys | 0.5h | Arrow keys, F1-F12, etc. |

### 2.4 Smart Form Fill
| Task | Effort | Açıklama |
|------|--------|----------|
| fillForm() command | 2h | {field: value} mapping |
| Auto-detect fields | 1h | name/id/label matching |
| Select/dropdown support | 1h | <select> elementleri |
| Checkbox/radio support | 1h | Boolean inputs |
| Date picker support | 1h | Date inputs |

**Phase 2 Toplam:** ~18 saat

---

## Phase 3: AI Integration (v0.3.0)

**Hedef:** 1 hafta  
**Öncelik:** Yüksek

### 3.1 OpenClaw Skill
| Task | Effort | Açıklama |
|------|--------|----------|
| SKILL.md | 1h | Kullanım dokümanı |
| TypeScript wrapper | 2h | API client |
| Command builders | 2h | click(), type(), screenshot() |
| Error handling | 1h | Retry logic, graceful errors |
| Examples | 1h | Örnek kullanımlar |

### 3.2 Natural Language Commands
| Task | Effort | Açıklama |
|------|--------|----------|
| Element description parsing | 3h | "mavi butona tıkla" → selector |
| LLM integration | 2h | GPT/Claude ile element bulma |
| Fallback to selector | 1h | NL başarısız olursa |
| Caching | 1h | Aynı element için cache |

### 3.3 Smart Actions
| Task | Effort | Açıklama |
|------|--------|----------|
| Auto-login | 2h | Saved credentials ile login |
| Cookie management | 2h | Import/export cookies |
| Session persistence | 2h | Login durumunu koru |

**Phase 3 Toplam:** ~20 saat

---

## Phase 4: Advanced Features (v0.4.0)

**Hedef:** 2 hafta  
**Öncelik:** Orta

### 4.1 Session Recording
| Task | Effort | Açıklama |
|------|--------|----------|
| Command history | 2h | Tüm komutları kaydet |
| Replay engine | 3h | Kaydedilmiş komutları tekrar çalıştır |
| Export/Import | 2h | JSON format |
| UI: Recording controls | 2h | Record/Stop/Play buttons |

### 4.2 Visual Selector
| Task | Effort | Açıklama |
|------|--------|----------|
| Click-to-select mode | 3h | Sayfada elemente tıkla → selector al |
| Highlight overlay | 2h | Seçili elementi vurgula |
| Selector suggestions | 2h | Multiple selector options |
| Copy to clipboard | 0.5h | Selector'ı kopyala |

### 4.3 iFrame Support
| Task | Effort | Açıklama |
|------|--------|----------|
| iFrame detection | 2h | Sayfadaki iframe'leri listele |
| Cross-origin handling | 3h | CORS issues |
| Nested iFrame | 2h | iframe içinde iframe |
| Frame switching | 1h | switchToFrame() command |

### 4.4 File Operations
| Task | Effort | Açıklama |
|------|--------|----------|
| File upload | 3h | input[type=file] |
| File download | 2h | Download trigger + track |
| Drag & drop files | 2h | Drag file to element |

**Phase 4 Toplam:** ~30 saat

---

## Phase 5: Platform Expansion (v0.5.0)

**Hedef:** 2-3 hafta  
**Öncelik:** Orta-Düşük

### 5.1 Firefox Extension
| Task | Effort | Açıklama |
|------|--------|----------|
| Manifest conversion | 2h | MV3 → Firefox format |
| API differences | 4h | chrome.* → browser.* |
| Testing | 3h | Firefox-specific issues |
| Firefox store publish | 2h | AMO submission |

### 5.2 Edge Extension
| Task | Effort | Açıklama |
|------|--------|----------|
| Edge compatibility | 2h | Chromium-based, minimal changes |
| Edge store publish | 1h | Microsoft Partner Center |

### 5.3 Safari Extension (Future)
| Task | Effort | Açıklama |
|------|--------|----------|
| Swift wrapper | 8h | Safari extension architecture |
| App Store | 4h | Apple review process |

**Phase 5 Toplam:** ~25 saat

---

## Phase 6: Enterprise Features (v1.0.0)

**Hedef:** 1 ay  
**Öncelik:** Düşük (talebe göre)

### 6.1 Team Management
- Multi-user support
- Role-based permissions
- Shared tokens
- Audit logging

### 6.2 Dashboard
- Web-based control panel
- Usage analytics
- Token management UI
- Real-time monitoring

### 6.3 Advanced Security
- 2FA for tokens
- IP whitelisting
- Custom blacklists
- Encryption at rest

### 6.4 High Availability
- Redis for state (optional)
- Horizontal scaling
- Load balancing
- Health monitoring

---

## Timeline Özeti

```
2026 Şubat
├── Week 2 (current)
│   └── ✅ v0.1.1 MVP + Hotfix
│
├── Week 3
│   ├── v0.2.0 Multi-tab + Wait + Keyboard + Form
│   └── v0.3.0 OpenClaw Skill + AI Commands
│
└── Week 4
    └── v0.4.0 Recording + Visual Selector + iFrame

2026 Mart
├── Week 1-2
│   └── v0.5.0 Firefox + Edge
│
└── Week 3-4
    └── v1.0.0 Enterprise (if needed)
```

---

## Prioritization Matrix

```
                    IMPACT
              Low    Med    High
         ┌─────────────────────────┐
    Low  │ Safari │ Edge  │Firefox │
         ├─────────────────────────┤
 EFFORT  │ Visual │Record │ iFrame │
    Med  │Selector│  ing  │        │
         ├─────────────────────────┤
    High │  Form  │ Wait  │Multi-  │
         │  Fill  │ Cond  │ Tab    │
         └─────────────────────────┘
         
         🎯 Start from bottom-right (High Impact, Low Effort)
```

---

## Success Metrics

| Metric | Target | Current |
|--------|--------|---------|
| Command latency | <500ms | ~200ms ✅ |
| Screenshot time | <2s | ~1s ✅ |
| Extension size | <500KB | ~50KB ✅ |
| Docker image | <20MB | ~10MB ✅ |
| Concurrent connections | 100+ | TBD |
| Uptime | 99.9% | TBD |

---

## Contributing

Katkıda bulunmak isteyenler için:
1. Roadmap'ten bir task seç
2. Issue aç
3. PR gönder

**Labels:**
- `good-first-issue` — Yeni başlayanlar için
- `help-wanted` — Yardım istenen
- `priority-high` — Öncelikli

---

*Bu roadmap yaşayan bir dokümandır. Öncelikler değişebilir.*
