# trackit
`trackit` is a cross-platform, light-weight CLI (command-line-interface) personal finance tracking tool. It's
meant for power-users (programmer types) who prefer working on the command-line over GUIs. It is mainly
for tracking transactions across accounts on a monthly basis.

Its main method of ingesting transaction records is by parsing and importing CSV files downloaded from your
bank accounts into an embedded, file-based [sqlite](https://sqlite.org/) database.

## Features
- [x] **Cost**: free.
- [x] **Cross-platform**: MacOS, Windows, Ubantu.
- [x] **Entirely offline**: no internet connection needed, and thus, no inherent privacy concerns.
- [x] **Multi-device**: leverages file-based database, so you are in charge of how
   (or if you even want to) sync the `*.db` file across multiple devices.
- [x] **Flexible**: you can write custom queries against the auto-generated `*.db` sqlite file.
- [x] **Performance**: It's all file-based and written in GO with embedded sqlite.
- [x] **Multi currency**: yes!
- [x] **Categorization**: tag with built-in categories, or manage your own custom categories.
- [x] **Ignore selected transactions**: Mark certain transactions ignored (such as transfers), so they don't get included in
      sums/aggregation.

## Getting started
1. [Download](https://github.com/kahunacohen/trackit/releases/) the correct version of trackit for your operating system
1. Put the `trackit` executable in your path.
1. The data directory (from which it imports CSV files from) is at `~/trackit-data`, but you can change that with the `--data-path` flag to
   `trackit init`.
1. Add date directories and CSV files. For example, create directory `~/trackit-data/2024-10` and download a month's transaction CSV file
   from your bank account for that month to that directory. Rename the CSV file after the name of your account (e.g. `bank_of_america.csv`). You should now have a file at: `~/trackit-data/2024-10/bank_of_america.csv`. Put other accounts' transaction CSV files for that month
   under that same directory.
1. Create a `trackit.yaml` configuration file. Each account key in the `trackit.yaml` file should match the corresponding file name under each
   month directory. In the `trackit.yaml` file, map each CSV heading for each account to one of the three required trackit database tables. These tables are:
   
   * `transaction_date`
   * `counter_party`
   * `amount`

If a header in the CSV file doesn't map to any table, set the table to: `~`. Here's an example `trackit.yaml` configuration file:

```yaml
accounts:
  bank_of_america:
    # the date reference layout for this account's CSV.
    # The reference date is Jan 1 2006. The entry below means
    # that for the bank_of_america CSV files, the date is formatted
    # mm/dd/yyyy.
    date_layout: 01/02/2006
    headers:
      - name: Posted Date # This is the CSV column header for the bank_of_america.csv file
        table: transaction_date # This is the trackit database table it maps to

      - name: Reference Number
        table: ~ # There is no trackit table for this column

      - name: Payee
        table: counter_party # counter_party is the subject party for the transaction

      - name: Address
        table: ~

      - name: Amount
        table: amount
```

By default the `trackit.yaml` file should be at `~/trackit.yaml`, but you can customize that in the `--config-file` flag
to `trackit init`.

After running `trackit init` (see above), run `trackit import`. Now you can list transactions (`trackit list`), search (`trackit search`)
etc. Run `trackit -h` for more options.

## Manual Transactions
You can manually add transactions (e.g. cash transactions) with `trackit add`. See `trackit add -h` for more.

## Multi-currency
trackit supports multi-currency. Add a `base_currency` key in `trackit.yaml` and set the appropriate currency code
for each account. E.g.

```yaml
base_currency: ILS # Israeli shekel
accounts:
  bank_of_america:
    date_layout: 01/02/2006
    currency: USD # Bank of America uses US dollars.
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

## Aggregating
You can aggregate transactions and get monthly reports by category using `trackit aggregate`. Currently
aggregating by other facets is not implemented. But you can run custom SQL queries.

## Ignoring transactions
Sometimes you might want to mark a transaction to ignore for summing or aggregation purposes. Say, for example, you
make a transfer from one account to another. You might not want trackit to see that as a debit or credit to your overall
spending. See `trackit ignore`.

## Custom queries/Syncing
Because the data is stored in a relational sqlite db (in the trackit.db file), you can make custom
queries against the database. You must install sqlite. Then do:

```
sqlite3 ~/trackit.db
```

...and make whatever queries you like. You can also save queries in a file and run them. You could
sync the db file using Google drive, or some other mechanism, to other devices and install trackit on those devices.

## More
For more about what you can do with `trackit`, see the help. E.g.

```
trackit -h
trackit init -h
trackit categorize -h
# etc.
```
