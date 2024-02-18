# Compute scheduler blog post

## CODE QUALITY NOTE:

There are lots of cases where I have `must` variations of functions that just panic on error, as well as many `logger.Fatal()` calls. Obviously this is a bad idea, and is entirely done to keep the code as terse as possible.

This code should not be forked for production use, but serve as a point of reference for how this system works.