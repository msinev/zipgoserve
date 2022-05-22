# zipgoserve
Very simple implementation for serving static file content from of ZIP archive. 
Serves compressed content without decompression where possible.
Assumes only deflate or store methods used in archive (e.g. standard linux zip utility)

## TO DO
 - Check if mutex is needed for zip file access (implement as option?)
 - Benchmark results
 - Preload small files to bytes slice (performance optimisation)
 - ZIP file as example
 - Dockerisation example
 - JSON mime maping example