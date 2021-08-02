(let x [1, 2, sum])

(puts "sum of 1, 2, 3 is ", (|x 2| [1, 2, 3]))

(let isStr (fn (x):
    (== (type x) "STRING")
    )
)

(let isStrArr (fn (arr):
    (map arr, isStr)
    )
)

(let concat (fn (arr):
    (reduce arr, "", incr)
    )
)

(let describeMe (fn (dict):
    (let name (+ |dict "name"| " is "))
    (let age (+ |dict "age"| " years old! "))
    (let likes (concat |dict "likes"|))
    (+ name (+ age (+ "He enjoys " likes)))
    )
)

(puts (describeMe {
    "name": "Joshua", 
    "age": "22", 
    "likes": 
        ["naps, ", 
        "cooking, ", 
        "reading, ", 
        "and bare other shiit"]}))

(let Ellispfullname (fn (f):
        (fn (m):
            (fn (l):
                (+ "your names are " (concat [f, m, l])))))
)

(putl (((Ellispfullname "joshua") " tamunoiwarilama") " pepple"),
        (((Nalafullname "joshua") " tamunoiwarilama") " pepple"))