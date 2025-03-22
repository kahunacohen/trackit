# trackit
`trackit` is a cross-platform, light-weight CLI (command-line-interface) personal finance tracking tool. It's
meant for power-users (programmer types) who prefer working on the command-line over GUIs and prefer to avoid
network traffic when managing personal financial data.

It's essentially a light wrapper over embedded [SQLite](https://sqlite.org/). SQLite is a fast, file-based database and behaves
like other relational databases (e.g. MySQL, Postgres etc.).

trackit's main method of ingesting transactions is by parsing and importing CSV files downloaded from your
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
1. Put the `trackit` executable in your path. You may be warned that it's not a trusted app on the Mac or Windows platform. I don't want to pay
   to be an Apple developer, so using the app is at your own risk. See [here](https://support.apple.com/en-il/guide/mac-help/mh40616/mac) if you want to override security settings and use it anyway.
   See [here](https://support.garmin.com/en-US/?faq=IWKKPZkMLD6dxzny6ksCK9) for windows.
1. Create the directory where trackit data will be stored. This is the directory where the `trackit.db` (the SQLite db file) database file, the trackit.yaml config file and
   your downloaded monthly CSV files will be stored. You can create this directory in your home directory:
   
   ```
   # Linux/MacOS
   mkdir $HOME/trackit-dir

   # Windows
   mkdir %USERPROFILE%\trackit-data
   ```

   You can create this directory anywhere you want and name it what you want, as long as it matches the environment
   variable you will set below. But in most cases, creating it at `$HOME/trackit-data` suffices.
1. Set the environment variable `TRACKIT_DATA` to the directory path you just created. In Linux/MacOS, put `export TRACKIT_DATA=$HOME/trackit-data`
   in your `.bashrc` or `.zshrc` file--then make sure to source it (`source ~/.bashrc`). In Windows run `setx TRACKIT_DATA "%USERPROFILE%\trackit-data"` to set the enviroment variable permanently. 
1. Download monthly transactions from your bank in CSV format and put them into your data directory. You can organize the 
   files however you like in that directory, **as long as the name of the file contains the bank account key** somewhere in the file name. The bank account key is the name of the bank account with underscores that you set in the `trackit.yaml` file below. For example if one of your bank account keys
   is `bank_of_america`, you can name the file `bank_of_america_transactions.csv`, as long as **bank_of_america** is a substring
   of the file name. **Note**: trackit can only import transactions from CSV files that are well-formed. Some banks will, unfortunately, download badly-formed CSV files. For example
   data rows may have unescaped quotes. It's your responsiblity to ensure the CSV files are clean, otherwise trackit will fail to import.
1. Create a `trackit.yaml` configuration file and put it in your trackit data directory you created earlier. 
   In your `trackit.yaml` file, map each CSV heading for each account to one of the three required trackit database tables. These tables are:
   
   * `transaction_date`
   * `counter_party`
   * `amount`

If a header in the CSV file doesn't map to any table, set the table to: `~`. Here's a simple, example `trackit.yaml` configuration file:

```yaml
accounts:
  bank_of_america: # this is the one of the bank acount keys: bank_of_america

    # the date reference layout for this account's CSV.
    date_layout: mm/dd/yyyy

    # These are the column headers of the bank_of_america CSV file, each mapped to the trackit database
    # tables.
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
> duplicate entries. The CSV files are not meant to be altered after importing.

> [!WARNING]  
> Although you can query the trackit SQLite database, and even manually add entities using SQL, don't
> modify the schema itself--that should only be managed by trackit. New versions of the trackit executable may
> perform schema migrations that could alter the schema.

## Manual Transactions
You can manually add transactions (e.g. cash transactions) with `trackit transaction create`. See `trackit transaction create -h` for more.

## Multi-currency
trackit supports multi-currency. Add a `base_currency` key in `trackit.yaml` and set the appropriate currency code
for each account. Here's an example of a more involved `trackit.yaml` file with multi-currency:

```yaml
base_currency: ILS # Israeli shekel
accounts:
  bank_of_america:
    date_layout: dd/mm/yyyy
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
  leumi_checking: # Here's a second account
    thousands_separator: ","
    currency: ILS
    date_layout: dd/mm/yyy
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
      # then map them to deposit/withdrawl tables instead of one amount table. E.g.:
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

Now, for each month, get the average conversion rate (perhaps look online) and add it with `trackit rate create`.

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
spending. See `trackit transaction ignore`.

## Working across machines
It's advisable to back up the `~/trackit-data` directory, as that's where your CSV files, `trackit.yaml` config file, and your `trackit.db` database file are located. You could, for example, manage that directory as a github repo and push/pull
it from multiple machines.

## Custom queries/Syncing
Because the data is stored in a relational SQLite db (in the trackit.db file), you can make custom
queries against the database. For now you must install sqlite on your platform (if it's not already). Then do:

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