# go-pgtesthelper
Small package allowing the use of disposable postgres dbs for testing purposes.

You don't want to use this.  Instead take a look at: https://github.com/viant/dsunit

<a name="Motivation"></a>

## Motivation
I have never liked relying on development data for unit tests.  Thankfully I once worked at a shop where, despite having hundreds of complex RDBMS tests, we never had to worry about them affecting our development data.  This was do to the forsight of the senior developer in the group deciding that for testing we should rely on ephemeral databases.

I liked that pattern and missed at a recent gig, so decided to take a stab at coming up with very simple Postgres-based approach.


