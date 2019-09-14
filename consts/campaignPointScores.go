package consts

var PointScores = map[int64][]int64{ // chapterNum: []int{...}
    //1: []int64{1000, 2000, 3000, 4000, 5000, -1}, // last element is the boss fight point, should never encounter
    //2: []int64{50000, 75000, 100000, 125000, -1},
    1: []int64{1000, 1000, 1000, 1000, 1000, -1}, // last element is the boss fight point, should never encounter
    2: []int64{50000, 50000, 50000, 50000, 50000, -1},
    3: []int64{50000, 50000, 50000, 50000, 50000, -1},
}
