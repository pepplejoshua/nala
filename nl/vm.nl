let prog = [
    "PUSH", 3,
    "PUSH", 4,
    "ADD",
    "PUSH", 5,
    "MINUS", 
    "PUSH", 20, 
    "MULTIPLY"
]

putl("..program statements..",prog)
putl()

let virtualMachine = fn(prog) {
    let pc = 0
    let stack = []
    let sp = 0
    inner(pc, stack, sp)
}

let inner = fn(pc, st, sp) {
    if (pc < len(prog)) {
        let curIns = prog[pc]
        
        if (type(curIns) == "INTEGER") {
            let pc = incr(pc, 1)
            return inner(pc, st, sp)
        }
        if (curIns == "PUSH") {
            let ind = pc+1
            ins(st, sp, prog[ind])
            let pc = incr(pc, 1)
            let sp = incr(sp, 1)
            return inner(pc, st, sp)
        }

        if (curIns == "ADD") {
            let r = st[sp-1]
            let sp = sp - 1

            let l = st[sp-1]
            let sp = sp - 1


            ins(st, sp, l + r)
            let sp = sp + 1
            let pc = pc + 1

            return inner(pc, st, sp)
        }
        if (curIns == "MINUS") {
            let r = st[sp-1]
            let sp = sp - 1

            let l = st[sp-1]
            let sp = sp - 1


            ins(st, sp, l - r)
            let sp = sp + 1
            let pc = pc + 1

            return inner(pc, st, sp)
        }
        if (curIns == "MULTIPLY") {
            let r = st[sp-1]
            let sp = sp - 1

            let l = st[sp-1]
            let sp = sp - 1


            ins(st, sp, l * r)
            let sp = sp + 1
            let pc = pc + 1

            return inner(pc, st, sp)
        }
    }
    puts("result=> ", st[sp-1])
}

virtualMachine(prog)