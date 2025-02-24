# trackit
`trackit` is a cross-platform, light-weight CLI (command-line-interface) personal finance tracking tool. It's
meant for power-users (programmer types) who prefer working on the command-line over GUIs and prefer to avoid
network traffic when managing personal financial data.

It's essentially a light wrapper over embedded [SQLite](https://sqlite.org/). SQLite is a fast, file-based database and behaves
like other relational databases (e.g. MySQL, Postgres etc.).

Its main method of ingesting transactions is by parsing and importing CSV files downloaded from your
bank accounts, but it also allows you to manually manage transactions.

## Features
- [x] **Cost**: free.
- [x] **Cross-platform**: MacOS, Windows, Ubantu.
- [x] **Entirely offline**: no internet connection needed, and thus, no inherent privacy concerns.
- [x] **Multi-device**: leverages file-based database, so you are in charge of how
   (or if you even want to) sync the `trackit.db` file across multiple devices.
- [x] **Flexible**: you can write custom queries against the auto-generated `trackit.db` sqlite file.
- [x] **Performance**: It's all file-based and written in GO. It's fast.
- [x] **Multi currency**: yes!
- [x] **Categorization**: tag transactions with built-in categories, or manage your own custom categories.
- [x] **Ignore selected transactions**: Mark certain transactions ignored (such as transfers), so they don't get included in
      sums/aggregations.

## Getting started
1. [Download](https://github.com/kahunacohen/trackit/releases/) the correct version of trackit for your operating system (under `assets`).
1. Unzip the zip file.
1. Put the `trackit` executable in your path.
1. The data directory (the directory where `trackit` imports CSV files from) is, by default, at `~/trackit-data`.
   Create that directory, if it doesn't exist. You can change the default data directory if you want it to be somewhere else
   on your filesystem. See `trackit init -h`.
1. Download monthly transactions from your bank in CSV format and put them into your data directory. You can organize the 
   files however you like in that directory, **as long as the name of the file contains the bank account key** somewhere in the file name. The bank account key is the name of the bank account with underscores that you set in the `trackit.yaml` file below. For example if one of your bank account keys
   is `bank_of_america`, you can name the file `bank_of_america_transactions.csv`, as long as **bank_of_america** is a substring
   of the filename.
1. Create a `trackit.yaml` configuration file. By default it will be `~/trackit-data/trackit.yaml`, but you can set a custom location.
   See `trackit init -h`. In your `trackit.yaml` file, map each CSV heading for each account to one of the three required trackit database tables. These tables are:
   
   * `transaction_date`
   * `counter_party`
   * `amount`

If a header in the CSV file doesn't map to any table, set the table to: `~`. Here's an example `trackit.yaml` configuration file:

```yaml
accounts:
  bank_of_america: # this is the one of the bank acount keys: bank_of_america

    # the date reference layout for this account's CSV.
    # The reference date is Jan 1 2006. The entry below means
    # that in the bank_of_america CSV files, the dates are formatted
    # like:  mm/dd/yyyy.
    date_layout: 01/02/2006

    # These are the headers of the bank_of_america CSV file
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

Now run `trackit transaction import`. That should import all transactions from your CSV files.

> [!WARNING]  
> If you edit a CSV file once it's been imported, trackit will see it as a new file and re-import all its rows--causing
> duplicate entries. The CSV files are meant to be an append-only log. Once imported, the files shouldn't be touched.

> [!WARNING]  
> Although you can query the trackit database all you like and even manually add entities using SQL, don't
> modify the schema itself--that should only be managed by trackit. New versions of the trackit executable may
> perform schema migrations that could alter the schema. Let trackit handle that.

## Manual Transactions
You can manually add transactions (e.g. cash transactions) with `trackit transaction create`. See `trackit transaction create -h` for more.

## Multi-currency
trackit supports multi-currency. Add a `base_currency` key in `trackit.yaml` and set the appropriate currency code
for each account. Here's an example of a more involved `trackit.yaml` file with multi-currency:

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

      # if the CSV file has separate withdrawl fields and deposit fields (instead of one amount field),
      # then map them to deposit/withdrawl table instead of one amount table.
      - name: withdrawl 
        table: withdrawl
      - name: deposit
        table: deposit
      - name: balance
        table: ~
      - name: remark
        table: ~
```

`trackit` comes populated with commonly used currency symbols. See `trackit currency list`. If you need to add a currency,
do `trackit currency create`. 

Now, for each month, get the average conversion rate and add it with `trackit rate create`.

## Separate deposit/withdrawl fields
Some downloaded CSV transaction rows have an amount field that is positive (deposits) or negative (withdrawls). Other
CSV downloads will have separate columns for deposits and withdrawls. `trackit` has an `amount` table for the former case
and `deposit`/`withdrawl` tables for the latter case. See example yaml above.

## Aggregating
You can view aggregate transactions and get monthly reports by category using `trackit transaction aggregate`. Currently
aggregating by other facets is not implemented. But you can run custom SQL queries.

## Ignoring transactions
Sometimes you might want to mark a transaction to ignore for summing or aggregation purposes. Say, for example, you
make a transfer from one account to another. You might not want trackit to see that as a debit or credit to your overall
spending. See `trackit ignore`.

## Working across machines
It's advisable to back up the `~/trackit-data` directory, as that's where (by default) all your CSV files, `trackit.yaml` config file, and your `trackit.db` database file are located. You could, for example, manage that directory as a github repo and push/pull
it from multiple machines.

If one one machine the data directory is at another location, you have to register the location of the data directory (CSV files), the `trackit.yaml` file and the database file, `trackit.db` using `trackit init`.

## Custom queries/Syncing
Because the data is stored in a relational sqlite db (in the trackit.db file), you can make custom
queries against the database. For now you must install sqlite on your platform. Then do:

```
sqlite3 ~/trackit.db
```

...and make whatever queries you like. You can also save queries in a file and run them:

```
sqlite3 ~/trackit.db < custom.sql
```

## More
For more about what you can do with `trackit`, see the help. E.g.

```
trackit -h
trackit init -h
trackit categorize -h
# etc.
```

## TODO
### Bugs
* settings table is allowing multiple rows with same name, table should have a unique constraint.
* This means you can have the wrong path to the config file. There needs to always be a check so that we don't have to call
  init every time we work across different machines.
  
### Features
1. Pre loop and create accounts before walking so we don't have to get the account ID
   in the middle of the transaction when walking.
1. Add category when categorizing
1. Tagging
1. Descriptions
1. Search across categories
1. Add unique constraint on name in accounts