# Theory
## LSEQ Positioning: 
In CRDTs for ordered structures like arrays or text, each element is assigned a **position** — a list of integers (e.g. `[5]`, `[10, 3]`). These positions are ordered **lexicographically**, allowing us to insert new elements between existing ones without rearranging them.

The `generatePositionBetweenLSEQ(left, right)` function generates a new position **between two given positions**.

CRDTs must support **concurrent inserts**. When two elements are inserted "at the same place," we can’t store both at the same position. LSEQ solves this by creating **new positions between existing ones**.
This is especially important in collaborative systems like:
- JSON CRDT arrays
- Real-time collaborative editors (e.g., inserting characters between `A` and `B`)

### How It Works
- Starts at level 0.
- Compares digits of `left` and `right`.
- If there's a **gap** (e.g., `left = 4`, `right = 8`), it picks a number in between (e.g., `6`).
- If **no room**, it goes deeper (e.g., `[4, 9, 9]` vs `[5]`) to create a unique sortable position.

### Examples

| #  | Left        | Right       | Result (Example)    | Notes                                    |
|----|-------------|-------------|---------------------|------------------------------------------|
| 1  | `[1]`       | `[4]`       | `[2]`               | Simple gap at level 0                    |
| 2  | `[2]`       | `[5]`       | `[3]` or `[4]`      | Multiple values between                  |
| 3  | `[3]`       | `[4]`       | `[3, x]`            | No gap → go deeper                       |
| 4  | `[4]`       | `[6]`       | `[5]`               | Single mid value                         | 
| 5  | `[]`        | `[3]`       | `[1]` or `[2]`      | Unbounded left → insert at beginning     |
| 6  | `[10]`      | `[]`        | `[11]`–`[15]`       | Unbounded right → insert at end          |
| 7  | `[15]`      | `[]`        | `[15, x]`           | Max digit reached → go deeper            |
| 8  | `[6]`       | `[6, 3]`    | `[6, 1]`, `[6, 2]`  | Inserting before deeper right            |
| 9  | `[6, 0]`    | `[6, 4]`    | `[6, 1]`–`[6, 3]`   | Simple split at level 1                  |
| 10 | `[5, 2]`    | `[5, 5]`    | `[5, 3]`            | Level 1 has room                         |
| 11 | `[5, 15]`   | `[6]`       | `[5, 15, x]`        | Level 1 maxed, go to level 2             |
| 12 | `[1, 7]`    | `[1, 9]`    | `[1, 8]`            | Room at level 1                          |
| 13 | `[0]`       | `[16]`      | `[1]`–`[15]`        | Full range possible                      |
| 14 | `[2]`       | `[2, 0]`    | `[2, x]`            | Must go deeper                           |
| 15 | `[2, 5]`    | `[2, 6]`    | `[2, 5, x]`         | No gap → level 2                         |
| 16 | `[8]`       | `[9]`       | `[8, x]`            | No room at level 0                       |
| 17 | `[6]`       | `[10]`      | `[7]`, `[8]`, `[9]` | Multiple simple options                  |
| 20 | `[7, 4]`    | `[7, 5]`    | `[7, 4, x]`         | Level 1 tight → insert at level 2        |

