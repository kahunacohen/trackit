# trackit
`trackit` is a cross-platform, light-weight CLI (command-line-interface) personal finance tracking tool. It's
meant for power-users (programmer types) who prefer working on the command-line over GUIs.

## Features
- [x] Cost: free.
- [x] Cross-platform: MacOS, Windows, Ubantu.
- [x] Entirely offline: no internet connection needed, and thus, no inherent privacy concerns.
- [x] Sycning: leverages file-based [sqlite](https://sqlite.org/) database, so you are in charge of how
   (or if you even want to) sync the `*.db` file across multiple devices.
- [x] Flexible: you can write custom queries because under-the-hood it uses sqlite.
- [x] Performance: fast (see above).
- [x] Multi currencies: yes!
- [x] Categorize transactions: tag with built-in categories, or manage your own custom categories.
- [x] Ignore selected transactions: Mark certain transactions (such as transfers), so they don't get included in
      sums/aggregation.

## Getting started
1. [Download](https://github.com/kahunacohen/trackit/releases/) the correct version of trackit for your operating system
1. Put the executable in your path.
1. By default, the data directory will be at `~/trackit-data`, but you can change that with the `--data-dir` flag to
   `trackit init`.
1. Add directories and CSV files. For example, create directory `~/trackit-data/2024-10` and donwload CSV file for
   that month to that directory. Rename the CSV file after the name of your account (e.g. `bank_of_america.csv`). You should now have a file at: `~/trackit-data/2024-10/bank_of_america.csv`.
1. Create a `trackit.yaml` configuration file. Each account key should match the corresponding file name under each
   month entry. Map each CSV heading to a database table  under `headings` in the trackit.yaml header. The table names
   in the database are:

   1. `transaction_date`
   1. `counter_party` (The party you either pay to or receive money from)
   1. `amount`

If a header in the CSV file doesn't map to any table, set the table to: `~`.

```yaml
accounts:
  bank_of_america:
    date_layout: 01/02/2006
    headers:
      - name: Posted Date
        table: transaction_date
      - name: Reference Number
        table: ~
      - name: Payee
        table: counter_party
      - name: Address
        table: ~
      - name: Amount
        table: amount
```

By default the config file should be at `~/trackit.yaml`, but you can customize that in the `--config-file` flag
to `trackit init`.

After running `trackit init` (see above), run `trackit import`. Now you can list transactions (`trackit list`), search (`trackit search`)
etc. Run `trackit -h` for more options.

## Multi-currency
trackit supports multi-currency. Add a base_currency key in `trackit.yaml` and set the appropriate currency code
for each account. E.g.

```yaml
base_currency: ILS
accounts:
  bank_of_america:
    date_layout: 01/02/2006
    currency: USD
    headers:
      - name: Posted Date
        table: transaction_date
      - name: Reference Number
        table: ~
      - name: Payee
        table: counter_party
      - name: Address
        table: ~
      - name: Amount
        table: amount
  leumi_checking:
    thousands_separator: ","
    currency: ILS
    date_layout: 02/01/2006
    headers:
      - name: date
        table: transaction_date
      - name: value date
        table: ~
      - name: description
        table: counter_party
      - name: reference
        table: ~
      - name: withdrawl
        table: withdrawl
      - name: deposit
        table: deposit
      - name: balance
        table: ~
      - name: remark
        table: ~
```

Then under each month directory, add a `rates.yaml` file. This should be set to the average exchange rate for
the month. E.g. at `~/trackit-data/2024-10/rates.yaml`, the yaml should look like:

```yaml
exchange_rates:
  - from: USD
    to: ILS
    rate: 3.75
```

For more about what you can do with `trackit`, see the help. E.g.

```
trackit init -h
trackit categorize -h
# etc.
```

## Custom queries/Syncing
Because the CSV files are parsed and stored in a relational sqlite db (in the trackit.db file), you can make custom
queries against the database. Install sqlite. Then do:

```
sqlite3 ~/trackit.db
```

and make whatever queries you like. You can also save queries in a file and run them. You can also
sync the db file using Google drive, or some other mechanism, to other devices and install trackit on those devices.

