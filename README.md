# zipgoserve
Very simple implementation for serving static file content from of ZIP archive. 
Serves compressed content without decompression where possible.
Assumes only deflate or store methods used in archive (e.g. standard linux zip utility)

## Benchmark
````
** SIEGE 4.0.4
** Preparing 25 concurrent users for battle.
The server is now under siege...
Lifting the server siege...
Transactions:		      817585 hits
Availability:		      100.00 %
Elapsed time:		       59.93 secs
Data transferred:	    14293.70 MB
Response time:		        0.00 secs
Transaction rate:	    13642.33 trans/sec
Throughput:		      238.51 MB/sec
Concurrency:		       23.98
Successful transactions:      817585
Failed transactions:	           0
Longest transaction:	        0.06
Shortest transaction:	        0.00
````


## TO DO
 - Check if mutex is needed for zip file access (implement as option  ✔)
 - Benchmark results  ✔
 - Preload small files to bytes slice (performance optimisation)  ✔
 - ZIP file as example
 - Dockerisation example
 - JSON mime maping example