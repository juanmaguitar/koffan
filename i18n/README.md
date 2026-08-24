# Translation System (i18n)

## Adding a New Language

### 1. Create a JSON File

Copy `en.json` as a template and create a new file, e.g., `de.json` for German:

```bash
cp i18n/en.json i18n/de.json
```

### 2. Edit Metadata

At the beginning of the file, modify the `meta` section:

```json
{
  "meta": {
    "code": "de",        // ISO 639-1 code (2 letters)
    "name": "Deutsch",   // Language name in that language
    "flag": "DE"         // Country code (displayed as flag)
  },
  ...
}
```

### 3. Translate All Keys

Translate values (NOT keys!) in each section:

- `common` - buttons: Add, Cancel, Save, Delete, Edit, Close
- `nav` - navigation: Settings, Logout
- `list` - list: title, empty list, shopping list, completed, Bought
- `items` - products: What to buy?, Note, name, new product, select section
- `sections` - sections: title, new section, section list, manage, select, no sections
- `actions` - actions: move up, move down, move, uncertain, certain
- `settings` - settings: title, language
- `login` - login: title, subtitle, password, placeholder, button, error
- `confirm` - confirmations: delete item, delete sections (with `{{name}}`, `{{count}}` parameters)

### 4. Rebuild the Application

JSON files are embedded in the binary, so after adding/changing translations:

```bash
go build -o shopping-list-go
./shopping-list-go
```

### 5. Done!

The new language will automatically appear in the Settings language selector.

---

## Translation File Structure

```json
{
  "meta": {
    "code": "xx",
    "name": "Language name",
    "flag": "XX"
  },
  "common": {
    "add": "...",
    "cancel": "...",
    "save": "...",
    "delete": "...",
    "edit": "...",
    "close": "..."
  },
  "nav": {
    "settings": "...",
    "logout": "..."
  },
  "list": {
    "title": "...",
    "empty_list": "...",
    "shopping_list": "...",
    "completed": "...",
    "bought": "..."
  },
  "items": {
    "what_to_buy": "...",
    "note": "...",
    "note_optional": "...",
    "name": "...",
    "new_product": "...",
    "select_section": "...",
    "section": "..."
  },
  "sections": {
    "title": "...",
    "new_section": "...",
    "section_list": "...",
    "manage": "...",
    "select": "...",
    "no_sections": "...",
    "add_first_section": "..."
  },
  "actions": {
    "move_up": "...",
    "move_down": "...",
    "move": "...",
    "uncertain": "...",
    "certain": "...",
    "remove_mark": "...",
    "mark_uncertain": "..."
  },
  "settings": {
    "title": "...",
    "language": "...",
    "coming_soon": "..."
  },
  "login": {
    "title": "...",
    "subtitle": "...",
    "password": "...",
    "password_placeholder": "...",
    "submit": "...",
    "error_invalid": "..."
  },
  "confirm": {
    "delete_item": "... \"{{name}}\"?",
    "delete_sections": "... {{count}} ...?",
    "delete_section": "... '{{name}}'?"
  }
}
```

## Language Codes (ISO 639-1)

| Code | Language |
|------|----------|
| pl | Polski |
| en | English |
| de | Deutsch |
| es | Español |
| fr | Français |
| it | Italiano |
| pt | Português |
| uk | Українська |
| cs | Čeština |
| sk | Slovenčina |
| ru | Русский |
| nl | Nederlands |
| sv | Svenska |
| no | Norsk |
| da | Dansk |
| fi | Suomi |
| ja | 日本語 |
| ko | 한국어 |
| zh | 中文 |

## Parameters in Translations

Some texts contain parameters in `{{param}}` format:

- `{{name}}` - element name (item, section)
- `{{count}}` - number of elements

Example:
```json
"delete_item": "Delete \"{{name}}\"?"
```

In JS code called as:
```javascript
t('confirm.delete_item', { name: 'Milk' })
// Result: Delete "Milk"?
```
