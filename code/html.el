(let p (fn (str):
    (+ (+ "<p>" str) "</p>")
))

(let h1 (fn (str):
    (+ (+ "<h1>" str) "</h1>")
))

(let (h1 str):
    (+ (+ "<h1>" str) "</h1>")
)

(putl (p "joshua pepple"))
(putl (h1 "A heading"))
