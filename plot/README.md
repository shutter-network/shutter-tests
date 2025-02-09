# plot hourly moving average of `shielded vs unshielded` inclusion rate

## Step 1 
Run psql query. This assumes this `README.md` and the scripts live in a directory called `plot/` and the directory above contains a `.env` file, that has an entry for `CONTINUOUS_DB_PASS=` (as is the case if you're running continuous tests).


    cd ..  # go to directory above 'plot/'
    docker run --rm -it \
        --entrypoint /usr/bin/psql \
        -e PGPASSWORD=$(cat .env|grep DB_PASS|cut -b20-) \
        --network observer_default \
        -v $(pwd)/plot:/plot \
        --workdir /plot \
        postgres \
            --csv \
            -f hourly_mov_avg.sql \
            postgres://postgres@db:5432/shutter_metrics


## Step 2
Create the plot. Run this from the directory above `plot/` as well:

    docker run --rm -it \
        --workdir /plot \
        -v $(pwd)/plot:/plot \
        remuslazar/gnuplot \
            -c gnuplot.script

Now you have a graph in `plot/plot.png`.
