let a = 1
let b = 2

let c = if (type(b) == "INTEGER") {
    a
} else {
    false
}

puts("a = ", a)
puts("b = ", b)
puts("c = ", c)

puts("c == a: ", c == a)
puts("b > a: ", b > a)
putl()

let userIn = reads("Enter something: ")
let likes = ["stuff", "wine", "code", "cooking meats", "reading"];
let name = "iwarilama";
let age = 22
let sex = "male"
let human = false

let me = {
    "name": name,
    "age": age, 
    "sex": sex,
    "isHuman": human,
    "enjoys": likes
}

puts("me[name]: ", me["name"])
puts(items(me))
putl()
puts(keys(me))
putl()
puts(values(me))
putl()

puts("you entered `", userIn, "` earlier!")
putl()

putl("summary:: ", me)

let genArrFromTypes = fn() {
    fn (arr) {
        map(arr, fn(x) { puts(x, "> ", type(x)) })
    }
}

let arr = [1, 2, 3, 4, type, sum, map, true, false, [1, 2, 3]]

genArrFromTypes()(arr)

puts(desc(type))

let t = fn(x) {
    let unt = 15
    let unt = incr(unt, 200)
    puts(unt)
}

t(1)