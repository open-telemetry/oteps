
## Generate a constants table

The constants table for the LookupTable histogram mapper is not
checked-in. The constants table can be used by any scale of histogram
less than or equal to the maximum scale computed.  To generate a
constants table, run:

```
go run ./generate MAXSCALE
```

For some value of `MAXSCALE`.  Note that the generated table will
contain 2**MAXSCALE entries.  Practical limits start to apply around
`MAXSCALE=16`, where this program takes days to run.
