# go-pgtesthelper
Small package allowing the use of disposable postgres dbs for testing purposes.

You don't want to use this.  Instead take a look at: https://github.com/viant/dsunit

<a name="Motivation"></a>

## Motivation
I have never liked relying on development data for unit tests.  Thankfully I once worked at a shop where, despite having hundreds of complex RDBMS tests, we never had to worry about them affecting our development data.  This was do to the foresight of the senior developer in the group deciding that for testing we should rely on ephemeral databases.

I liked that pattern and missed at a recent gig, so decided to take a stab at coming up with very simple Postgres-based approach.

<a name="Usage"></a>

## Usage
```
// Get a helper type value to work with
h, err := pgtesthelper.NewHelper("./sql/mydb.sql", "mydb", "myuser", "password", false)
if err != nil {
    return err
}

// Create a temporary database
dbh, err := h.CreateTempDB()
if err != nil {
    return err
}

// Get a reference to the handle of the just created database and query away
dbh := h.TestDB()
rows, err := dbh.Query(`
  INSERT INTO books (id, title, isbn, created_at) 
  VALUES ('cb0b9721-7631-4b2a-94a2-493c559da893','titleA', '9783161484100', NOW());`)
rows.Close()

// Do some testing

// Call CleanUp() to dispose of the database.
h.CleanUp()
```

Note that the call to CleanUp() could be a noop, meaning that the database won't be removed, if a value of true was passed as the last parameter (keepDB) to pgtesthelper.NewHelper().

Obviously, the user credentials passed into pgtesthelper.NewHelper() will need to have sufficient privileges on the 'postgres' database in order to create and drop the temporary databases.

Setting that up is left as an exercise to user.
