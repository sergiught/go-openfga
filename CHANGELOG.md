# Changelog

## 0.110.0 (2026-07-12)


### ⚠ BREAKING CHANGES

* **openfga:** service methods return (result, error) instead of (result, *Response, error); use OnResponse for raw response access.
* **models:** add strongly-typed model schema and DSL-aligned builders
* **config:** load FGA_* via os.Getenv, dropping caarlos0/env
* **models:** ReadModelsOptions and ReadAuthorizationModelsResponse are renamed to ListModelsOptions and ListAuthorizationModelsResponse.

### Features

* add chunking/parallelism request options and bounded worker helper ([3265164](https://github.com/sergiught/go-openfga/commit/3265164827644e9fdda69c819a6247cf5e892b01))
* add WithDefaultConsistency client option ([f6261b2](https://github.com/sergiught/go-openfga/commit/f6261b2e97db95af873e94c72bf7ffb8e390d0c3))
* API-token and OAuth2 client-credentials auth ([2f2144e](https://github.com/sergiught/go-openfga/commit/2f2144e7548ad3c5fd06b1aae687442c46ad3059))
* AssertionsService (write/read) ([815e137](https://github.com/sergiught/go-openfga/commit/815e137e267d8a293d1fd33f6ed6ef6e0f1e9cd2))
* AuthorizationModelsService (write/list/get/read-latest) ([c2027dc](https://github.com/sergiught/go-openfga/commit/c2027dcf306315f3b931baabc73101659c183f8e))
* auto-paginating iterators (Go 1.23 range-over-func) ([c202b57](https://github.com/sergiught/go-openfga/commit/c202b57c238e258246b03d3b82c07debf0571f4a))
* Client constructor, options, and service wiring ([1f2accc](https://github.com/sergiught/go-openfga/commit/1f2accc5adf563f367602fa4aeb9eff8b3be572e))
* **client:** composable transport, custom token source, and observer ([3632ded](https://github.com/sergiught/go-openfga/commit/3632dedf099acb8d34bf53d451c5a421bf3d48b3))
* **client:** ergonomic helpers, typed responses, and error context ([3b7d88e](https://github.com/sergiught/go-openfga/commit/3b7d88e65456eb0453f39096a220e15bcc38d1ff))
* **client:** make env loading opt-in via NewClientFromEnv/EnvOptions ([87b2179](https://github.com/sergiught/go-openfga/commit/87b21791a38f3ede1bd28f14c431c564ca0a4ff9))
* **client:** seed NewClient from FGA_* environment via config.Load ([373e940](https://github.com/sergiught/go-openfga/commit/373e940cc58eb3b6c4828d3461ea7202a458ca09))
* **client:** validate merged config once in NewClient ([f5785e1](https://github.com/sergiught/go-openfga/commit/f5785e1f161e2e623623af0886d33172c4fa0efa))
* **client:** version the default User-Agent and add default setters/getters ([f79f3f3](https://github.com/sergiught/go-openfga/commit/f79f3f38a00d43f92787543bae325feb447271cf))
* **config:** add internal/config with env-decoding Load ([4878dff](https://github.com/sergiught/go-openfga/commit/4878dff4c34708817c050c0049d003a4dd07fd69))
* **config:** load FGA_* via os.Getenv, dropping caarlos0/env ([ee2e6a3](https://github.com/sergiught/go-openfga/commit/ee2e6a337e8bcfe0e1f4761e77224ea9e9c29e60))
* **config:** normalize token issuer into an OAuth2 token URL ([1120f5a](https://github.com/sergiught/go-openfga/commit/1120f5a2ed2f0e204f0095114980b83fee180c43))
* Do and BareDo request execution ([5ce2d04](https://github.com/sergiught/go-openfga/commit/5ce2d04e1015a147b96f7d61a295f5f490bddefe))
* **dsl:** add DSL&lt;-&gt;JSON model transformer as a nested module ([7a8051e](https://github.com/sergiught/go-openfga/commit/7a8051e6c2753c5f954ea529593c4111053bbd86))
* expose Client.BaseURL ([3334017](https://github.com/sergiught/go-openfga/commit/3334017eb44c404cf805a2d571ea98138b82c61f))
* **models:** add strongly-typed model schema and DSL-aligned builders ([87b8e3a](https://github.com/sergiught/go-openfga/commit/87b8e3adb623b483594704b721fd4466b5fb5310))
* **openfga:** return (result, error) and add OnResponse ([6307e4c](https://github.com/sergiught/go-openfga/commit/6307e4c62f3f84924576981bbbbcf5beee5b0c42))
* private-key JWT (RFC 7523) auth ([6e8c1b2](https://github.com/sergiught/go-openfga/commit/6e8c1b2166b8f50c546d1ae717bf7607e422e84f))
* **relationships:** add client-side BatchCheckAll chunking helper ([2f06def](https://github.com/sergiught/go-openfga/commit/2f06def05a3cf8cde477e728579bd65612736bf6))
* **relationships:** add ListRelations built on BatchCheckAll ([d37f773](https://github.com/sergiught/go-openfga/commit/d37f773f988a48e5b678ad98cbf26629f6cd3901))
* **relationships:** return partial results from BatchCheckAll on failure ([f905023](https://github.com/sergiught/go-openfga/commit/f9050237c09fd3d28d07f42d229b0490ab4a85f9))
* RelationshipsService (check/batch-check/expand/list-objects/list-users) ([486506d](https://github.com/sergiught/go-openfga/commit/486506d678ffd77a73a525f4d387567d6069c2fb))
* request core (NewRequest, Response, request options) ([4f4c57f](https://github.com/sergiught/go-openfga/commit/4f4c57f5cc72e96a5484a3b745d495dccb55ea47))
* **response:** add QueryDuration and time-boxed fuzzing ([c6b1387](https://github.com/sergiught/go-openfga/commit/c6b138716ad4ece39c66d3271b1bfbd701bd4edc))
* retry transport (429 default, backoff, Retry-After) ([d8090f0](https://github.com/sergiught/go-openfga/commit/d8090f08dc70abbeb1d8579ce2d6f4be9c5cfacc))
* static-header transport ([86d67a2](https://github.com/sergiught/go-openfga/commit/86d67a20ab5fa3c049ef4acf630902537c7df628))
* StoresService (create/list/get/delete) ([c881264](https://github.com/sergiught/go-openfga/commit/c881264e99ea9a74ce8a50367bf7c5f8b1ac5c4e))
* **stores:** support the name filter in ListStores ([73ce293](https://github.com/sergiught/go-openfga/commit/73ce293a3dce3c437def01b4722cabaeb8c507ca))
* StreamedListObjects NDJSON iterator ([e6feb6a](https://github.com/sergiught/go-openfga/commit/e6feb6a9f6cad002dbcd4befe0b027243fa1d5ce))
* **tuples:** add bulk WriteTuples/DeleteTuples with per-tuple results ([4e1f425](https://github.com/sergiught/go-openfga/commit/4e1f425925e375cb0e2bb8ac27f4311834e35503))
* **tuples:** add on_duplicate/on_missing write-conflict options ([63f035f](https://github.com/sergiught/go-openfga/commit/63f035f131b77ee737d7ab43da1259029606abdf))
* **tuples:** report bulk partial failures via Failed and FirstError ([7cc9bb2](https://github.com/sergiught/go-openfga/commit/7cc9bb27d959fa95e7ab29f96f55c846a749b5a8))
* TuplesService (write/read/changes) ([259e669](https://github.com/sergiught/go-openfga/commit/259e6699eac95f5633dafbbf69650d61a3cebec7))
* typed API errors and response classifier ([4906a67](https://github.com/sergiught/go-openfga/commit/4906a67539e4f9c68576821edc64d13c3873667c))


### Bug fixes

* address foundation code-review (errors.As Unwrap, JWT status check, retry request clone, gofmt, tests) ([6d7321e](https://github.com/sergiught/go-openfga/commit/6d7321e53d74cee64b37073233c10a1276bfa245))
* avoid mutating caller request structs in TuplesService; test start_time ([ec8a798](https://github.com/sergiught/go-openfga/commit/ec8a798082b98e363b0a4b6b771bccdeb60891fb))
* cancellable retry backoff, WithHTTPClient docs, godoc examples ([7a0ccbc](https://github.com/sergiught/go-openfga/commit/7a0ccbc1706e4152c6487ebd08fcbfd1ac746e7f))
* **client:** drain nil-body responses so connections can be reused ([16e90bf](https://github.com/sergiught/go-openfga/commit/16e90bf53e104ab630dbb0a822c40a9a0c76ba19))
* **client:** harden retry and header transports ([0a7ffa3](https://github.com/sergiught/go-openfga/commit/0a7ffa388f3f3f5641ccd71832a9c77ebdc0c97a))
* encode FGAObjectRelation as OpenFGA's structured object ([f162cce](https://github.com/sergiught/go-openfga/commit/f162ccec13225240bd13f3886eaa5b41f646d80a))
* **errors:** guard Retry-After against negative and overflowing values ([bbaf5d4](https://github.com/sergiught/go-openfga/commit/bbaf5d4346142d56810e3afd35a7f822d947b4f0))
* **relationships:** use valid correlation IDs in ListRelations ([ec37fc3](https://github.com/sergiught/go-openfga/commit/ec37fc325b2457bb66991b58ca12c3f2c5334049))
* stop ChangesAll once the changes feed is drained ([833ee9f](https://github.com/sergiught/go-openfga/commit/833ee9f15fec4ebc6b0aa89d7c9b0a204077a0e7))


### Refactors

* **auth:** store credential specs and build transport in NewClient ([4fc1f56](https://github.com/sergiught/go-openfga/commit/4fc1f56a176da7e609aedc886ba9f84d495dbb50))
* **client:** layer env as options and extract NewClient assembly phases ([3c0642b](https://github.com/sergiught/go-openfga/commit/3c0642b66884525f077a320a51509f63df19fa3b))
* **client:** split validate into URL, ID, and retry checks ([1cf3cf6](https://github.com/sergiught/go-openfga/commit/1cf3cf6ac34449efd1be83f6f8ab9842e8d0f883))
* **models:** rename Read* model types to the List* convention ([f874dc7](https://github.com/sergiught/go-openfga/commit/f874dc7255cb4d3d591ed69d213ab4d6bf9d4a87))


### Tests

* add Expand and ListUsers runtime tests ([014f06b](https://github.com/sergiught/go-openfga/commit/014f06b1da013b8849c8ae8083294638b13d1002))
* cover AuthorizationModels.All and Tuples.ChangesAll iterators ([ed7ab92](https://github.com/sergiught/go-openfga/commit/ed7ab92b737119e49860f501758a0644e4c4a7d8))
* godog check scenarios against live OpenFGA ([119d8be](https://github.com/sergiught/go-openfga/commit/119d8beab6cf197cc0ccb4f83141e0b3d02550a6))
* implement OPENFGA_IMAGE override and drop dead suite state ([0c1f23d](https://github.com/sergiught/go-openfga/commit/0c1f23d11b906acff117cdb83f947225cde659f2))
* integration module scaffold with testcontainers OpenFGA ([3ace3fd](https://github.com/sergiught/go-openfga/commit/3ace3fd76667b729dc39f64e2118e1ddc47d928e))
* **integration:** add shared harness scaffolding and migrate check scenarios ([77d329a](https://github.com/sergiught/go-openfga/commit/77d329a81a4a7a2fe570e4fb442e4e0821814699))
* **integration:** cover assertions write and read round-trip ([2f267ff](https://github.com/sergiught/go-openfga/commit/2f267ffe90adda015ba78eda51e1fdf068919074))
* **integration:** cover authorization model write, get, read-latest, list, validation ([1f507eb](https://github.com/sergiught/go-openfga/commit/1f507eb0bd61ca4105f0cd9a14473977f456cd79))
* **integration:** cover batch-check, expand, list-objects, streamed-list-objects ([61cbfff](https://github.com/sergiught/go-openfga/commit/61cbfff8464cd265ae3445fa8dcebefd5e368b6e))
* **integration:** cover check inheritance, group membership, contextual tuples ([0d94dc2](https://github.com/sergiught/go-openfga/commit/0d94dc2c5999751db184cd3bb209d20db378f445))
* **integration:** cover list-users ([448bf6c](https://github.com/sergiught/go-openfga/commit/448bf6c95bf07c60070a1520aacd307fab9b54b0))
* **integration:** cover ListRelations and bulk write/delete chunking ([76bf968](https://github.com/sergiught/go-openfga/commit/76bf968cd2106e34e5a000ef00d930e3892a3615))
* **integration:** cover per-request overrides, consistency, headers, unbound-store error ([6eb13ed](https://github.com/sergiught/go-openfga/commit/6eb13ed41f38e71f21546bca1cb08a56e6bfbcfa))
* **integration:** cover store create, get, list-all, delete ([d618021](https://github.com/sergiught/go-openfga/commit/d61802127314f360df2c3cd2b00406b88852121f))
* **integration:** cover tuple write, read, delete, changes, validation ([8ec9a31](https://github.com/sergiught/go-openfga/commit/8ec9a3128a47f362bf19fa775a3f1a12a70a8d8b))
* **openfga:** add fuzz targets for response, codec, and header parsing ([17819d7](https://github.com/sergiught/go-openfga/commit/17819d7f0b215e25e713be57035efd060f0affee))
* **openfga:** cover error types, options, JWT token, and unmarshal paths ([45d2887](https://github.com/sergiught/go-openfga/commit/45d2887bc4f6ad2fe0819776e493e1102c9d7afa))
* use valid ULIDs for IDs passed through NewClient ([4f291f4](https://github.com/sergiught/go-openfga/commit/4f291f4fd866283adf77d88209bb88ac4b6bedd7))


### Documentation

* add contributing guide, code of conduct and security policy ([7e3c507](https://github.com/sergiught/go-openfga/commit/7e3c5070919d570fa33e1281aa1b99993804d3dc))
* add pull request and issue templates ([c57b25e](https://github.com/sergiught/go-openfga/commit/c57b25ec89cf01233d91d6d7e28469722b62b85b))
* add README ([17a2843](https://github.com/sergiught/go-openfga/commit/17a28430123ea689d1f2396714f20065eaba359e))
* add README banner and drop retired Go Report Card badge ([#6](https://github.com/sergiught/go-openfga/issues/6)) ([3ebb292](https://github.com/sergiught/go-openfga/commit/3ebb2921b344510eb8437482d04f579f8e8c1112))
* add table of contents, section emojis, and query/bootstrap/ABAC examples to README ([42a95c0](https://github.com/sergiught/go-openfga/commit/42a95c0c0ca83e8385322df8116fbe5ef0097a96))
* center the badge row in README ([301e89d](https://github.com/sergiught/go-openfga/commit/301e89d3b6c0f7ef507900150b6085f1742f219b))
* center the go-openfga title in README ([22f604a](https://github.com/sergiught/go-openfga/commit/22f604a04821465fd125b6126937543a366e5fa1))
* cover env config, errors, extensibility, and testing ([f74a79b](https://github.com/sergiught/go-openfga/commit/f74a79baa0ba6e722719d68f2716cb6286d7036f))
* document bulk helpers, batch-check chunking, conflict options, and dsl module ([3b31b73](https://github.com/sergiught/go-openfga/commit/3b31b73792e663507dc5a91ea742876d1ba74b2f))
* document dsl module layout and two-module release ordering ([1be6713](https://github.com/sergiught/go-openfga/commit/1be6713dce1cc853c2b094c4b7d7980fc39e8bca))
* document typed model authoring and bulk partial-failure helpers ([e3e5f79](https://github.com/sergiught/go-openfga/commit/e3e5f794fc4df984c8b2696ad95c3508d8956011))
* drop the go-github design comparison from the README intro ([d9532a9](https://github.com/sergiught/go-openfga/commit/d9532a931fdffad55b9dee4c6640857db4c27981))
* expand README badges and add NOTICE ([9239ab7](https://github.com/sergiught/go-openfga/commit/9239ab791522356bc2dec5bfc21dab7518910d58))
* update the code of conduct reporting link ([eb99831](https://github.com/sergiught/go-openfga/commit/eb998316e8037c13aaeac4da0da3dea61e10e6b3))


### Build

* add golangci-lint, commitlint, editorconfig and pre-commit config ([f0ed411](https://github.com/sergiught/go-openfga/commit/f0ed411a05437479a348654bf00fe534e138d1de))
* expand Makefile with lint, coverage, vuln and tooling targets ([4272e29](https://github.com/sergiught/go-openfga/commit/4272e290b1aedd30578a21b8f6456dcf343cd07c))
* **integration:** update test harness dependencies ([7f42479](https://github.com/sergiught/go-openfga/commit/7f42479bbfab615e4fd54540c8dd89653a0434b9))
* raise Go floor to 1.25 and update dependencies ([1795398](https://github.com/sergiught/go-openfga/commit/1795398314c964338cb3e9060197697172c5d4c3))


### Chores

* release core v0.110.0 ([b313e4e](https://github.com/sergiught/go-openfga/commit/b313e4e42fac4972b299a19eaada877e0259f831))
* release v0.38.0 ([a215762](https://github.com/sergiught/go-openfga/commit/a215762fadc1e1c43879c37f370611947b709b9f))

## Changelog

All notable changes to this project are documented in this file.

Releases and the entries below are generated by
[release-please](https://github.com/googleapis/release-please) from
[Conventional Commits](https://www.conventionalcommits.org). Until `v1.0.0` the
public API may change between minor versions.
