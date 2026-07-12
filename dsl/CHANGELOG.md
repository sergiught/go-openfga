# Changelog

## 1.0.0 (2026-07-12)


### ⚠ BREAKING CHANGES

* **openfga:** service methods return (result, error) instead of (result, *Response, error); use OnResponse for raw response access.
* **models:** add strongly-typed model schema and DSL-aligned builders
* **config:** load FGA_* via os.Getenv, dropping caarlos0/env

### Features

* **config:** load FGA_* via os.Getenv, dropping caarlos0/env ([ee2e6a3](https://github.com/sergiught/go-openfga/commit/ee2e6a337e8bcfe0e1f4761e77224ea9e9c29e60))
* **dsl:** add DSL&lt;-&gt;JSON model transformer as a nested module ([7a8051e](https://github.com/sergiught/go-openfga/commit/7a8051e6c2753c5f954ea529593c4111053bbd86))
* **models:** add strongly-typed model schema and DSL-aligned builders ([87b8e3a](https://github.com/sergiught/go-openfga/commit/87b8e3adb623b483594704b721fd4466b5fb5310))
* **openfga:** return (result, error) and add OnResponse ([6307e4c](https://github.com/sergiught/go-openfga/commit/6307e4c62f3f84924576981bbbbcf5beee5b0c42))
