gnuplot -persist <<-EOFMarker
    set style line 1 \
        linecolor rgb '#eb4034' \
        linetype 1 linewidth 1 \
        pointtype 7 pointsize 1.5
    set style line 2 \
        linecolor rgb '#67eb34' \
        linetype 1 linewidth 1 \
        pointtype 7 pointsize 1.5
    set style line 3 \
        linecolor rgb '#348feb' \
        linetype 1 linewidth 1 \
        pointtype 7 pointsize 1.5
    set style line 4 \
        linecolor rgb '#eb34e1' \
        linetype 1 linewidth 1 \
        pointtype 7 pointsize 1.5
    set style line 5 \
        linecolor rgb '#dd181f' \
        linetype 1 linewidth 1 \
        pointtype 5 pointsize 1.5
    set xlabel 'Identificador de Peticion, Linea Temporal de Ejecucion'
    set ylabel 'Tiempo de Ejecucion (segundos)'
    f(x) = 2.3 
    plot "secuencial.txt" using 1:2 title 'secuencial' with lines linestyle 1 , \
       "multithread.txt" using 1:2 title 'multithread' with lines linestyle 2 , \
       "poolthread.txt" using 1:2 title 'poolthread' with lines linestyle 3 , \
       "masterWorker.txt" using 1:2 title 'masterWorker' with lines linestyle 4, \
        f(x) title 'QoS deadline' with lines linestyle 5
EOFMarker
