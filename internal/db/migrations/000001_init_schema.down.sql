-- Drop the view first since it depends on transactions, accounts, and categories
drop view if exists transactions_view;

-- Drop tables in reverse order of creation to maintain foreign key integrity
drop table if exists rates;
drop table if exists currency_codes;
drop table if exists transactions;
drop table if exists categories;
drop table if exists accounts;
drop table if exists files;
drop table if exists settings;
