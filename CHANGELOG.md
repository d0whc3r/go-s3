## 1.1.2 (2020-07-24)

#### Chores

* **release:** fix remove temporal changelog (262f197b)
* **bump:** version [skip ci] (5acef2d5)

#### Bug Fixes

* get bucket from environment (f53035d1)

## 1.1.1 (2020-07-23)

#### Chores

* **deps:** update module aws/aws-sdk-go to v1.33.10 (ce6c363e)
* **release:** parse new changelog (927dad7f)
* **ci:** remove comments and unnecessary actions (666d2399)
* **make:** set version file to update (60f6bdf2)
* **deps:** add renovate.json (127f31da)
* **bump:** version [skip ci] (0b1bc842)

#### Bug Fixes

* **deps:** clean dependencies (5c0e73fe)

## 1.1.0 (2020-07-23)

#### Chores

* **make:** fix ginkgo path (d669d534)
* **ci:** stop and start mysql container (451a2de6)
* **ci:** stop and start mysql service for tests (72e8a116)
* **ci:** try other configuration for mysql volumes (c72840bc)
* **ci:** update mysql service init (032b05fb)
* **ci:** update mysql service (c9874bf1)
* **docker:** remove mysql start command (99b77fc4)
* **ci:** update env for mysql ci (e43a3616)
* **ci:** update ci with mysql service (679118b0)
* **coverage:** create coverage folder (c5a24d4a)
* **sql:** change location of sample sql file (0803dfb4)
* **sonar:** coverage files (3a11d2ef)
* **make:** update test process (7bc46c6f)
* **deps:** update dependencies (0814fec4)
* **go:** update go.sum (ec4fe265)
* **make:** update build names (6dae149d)
* **deps:** update dependencies (5c700312)
* **sonar:** config (41754bf3)
* **sonar:** fix sonar config (82b03524)
* **sonar:** configure sonar (c1b26121)
* **version:** bump version [skip ci] (39a6a214)
* **release:** update release script (fa51b468)
* **changelog:** update changelog (f0ee5c8e)
* **bump:** version [skip ci] (e942c477)

#### Documentation

* **readme:** update command output (1bf1a47c)
* **readme:** update github actions badge (665a39f4)
* add sonar coverage (cb46a04e)

#### Feature

* **cli:** add commands to cli (aaff7b26)
* **clean:** add clean old files function (765e18ef)

#### Bug Fixes

* **cmd:** set commands priorities (76e55332)
* **cmd:** logic in arg params (278913ba)
* **cmd:** reset options pointers in every exec (0352a8f9)
* **config:** empty default value (9bde103d)
* **wrapper:** check for bucket and endpoint (08f52c68)
* **version:** update version file (af03f480)
* **wrapper:** little bugs (227c92a4)
* expose DefaultUploadOptions (6936e647)
* **sonar:** code smell (5527d253)
* **sonar:** sonar issues (b4b79617)

#### Code Refactoring

* **split:** split code in multiple files (d23d29c2)
* huge refactor (3f69545e)

#### Styles

* use spaces (a9fae7c9)

#### Tests

* **cmd:** complete tests for command line (4a47c38e)
* **wrapper:** complete tests (90c717e5)
* **cmd:** basic tests (bf9fc56d)
* **wrapper:** add tests for wrapper (8d7f102d)
* remove commented code (8bb241c3)
* **config:** update tests using new library (6d62b377)
* **gos3:** update tests using new library (aac7cdc9)
* update test files for test tools (5d76b514)
* **tools:** create tools files for tests (f07f9fd5)
* add tests for all functions (52b66608)
* **docker:** add docker compose for local testing (2455f750)
* **env:** add test env file example (cba98edd)
* **makefile:** update test job (1f2a0ca6)
* update coverage (24f011b9)

## 1.0.3 (2020-07-20)

#### Chores

* **release:** update release functions (f44bc5af)
* **bump:** bump version [skip ci] (c8e103ef)
* **release:** set release git info (2c252de3)

#### Bug Fixes

* **docker:** update docker build (d1e14302)

## 1.0.2 (2020-07-20)

#### Chores

* **version:** release version (bd3ca37b)

#### Bug Fixes

* **release:** add git commit & push (3e438612)

## 1.0.1 (2020-07-20)

#### Chores

* **git:** ignore sonar folder (284acbdc)

#### Bug Fixes

* **sonar:** sonar issues (52b2d7e4)

## 1.0.0 (2020-07-20)

#### Chores

* **ci:** update pipeline (d94e37f6)
* **release:** use makefile for build (eaeada65)
* **ci:** add sonar (7d5b08ac)
* **version:** add version file (7b155b7e)
* **git:** update gitignore (629d5a93)
* **ci:** configure github action for go (4f8dff5d)
* **docker:** add docker (1b2ed7d8)
* **makefile:** add generic Makefile (d3aba649)
* **git:** add .gitignore (74f9bb17)

#### doc

* fix related projects (b0705212)
* add badges in readme (d5f41063)
* add Readme (790327e1)

#### Feature

* **main:** add main wrapper (e34cee49)
* **files:** add files (objects) management (013f2275)
* **config:** add configuration for env variables (6f8e3c04)
* **buckets:** add buckets management (1f207235)

#### Bug Fixes

* **release:** add release feature (51ef77a0)
* **shared:** add shared files functions (561abd95)
* **shared:** add shared buckets functions (883d7fd7)
* **cmd:** add initial cmd commands (390834f3)
* **module:** define go module (f0f480e0)

#### Tests

* add coverage in test (37882af5)
* add main tests (883097fd)

#### tests

* tests env file (9158e4ae)

