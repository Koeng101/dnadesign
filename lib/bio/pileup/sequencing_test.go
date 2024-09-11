package pileup

//var sequencingResults = map[string]SequencingResult{
//	"2eZetMG1kKOW3Iba84G6CPuaadY": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed\nConfirmed"},
//	"2eZetNv8fYHCM0DuBcLASMthXkp": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed\nConfirmed"},
//	"2eZetnAB4zyd7OVwnKTOfPVSzsu": {Confirmed: false, MixedTemplate: true, MixedColony: true, Notes: "This is a mixed colony with 2 different DNA constructs, with what appears to be different templates."},
//	"2eZeu0norbqCHxJVulXdyae1BNR": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed\nConfirmed"},
//	"2eZeu1JFNweG1wWffGt1Nmt35DP": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed\nConfirmed"},
//	"2eZeuUEOMFgF9zjssVEnEvASo4E": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed. 165 A->G nanopore sequencing error."},
//	"2eZeuPeZYKQY5vQiy9aAVbPsD3O": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "Mutated. 535 G->T. Looks like a mixed colony, but same template. NOT a nanopore sequencing error."},
//	"2eZeuagq6UhhV3X7NJVMfqENwmt": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed"},
//	"2eZeubvK0ttEByuJJUr5gA9VwfB": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed\nConfirmed"},
//	"2eZevBHl6NlvxG1URpizAQnyzYC": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed\nConfirmed"},
//	"2eZevLPVdxVQa1pbTl319XNTOG6": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed\nConfirmed"},
//	"2eZevu6eS97Kr0ocTimNGwe967v": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed.\n\n425 A->G is incorrect nanopore sequencing. You can tell because there is only one strand that has it."},
//	"2eZew5PGfoLNiVhoIiGqJkN8lNS": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed. 165 A->G is a nanopore error"},
//	"2eZexpXBtWXvSGyVVmY0NdQCre9": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed. 165 A->G is a nanopore error."},
//	"2eZeysCXdch5n3gCMkq1XUGmqcw": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed.\n\n165 A->G. Nanopore sequencing error"},
//	"2eZf0GO18rmfJyiSaOZ69lzMihN": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed.\n\n165 A->G. Nanopore sequencing error"},
//	"2eZf0K6koJQK8IjRU5MK7Qzunkg": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "mutations 263 C->T, 809 G->T. Note there is a lil of \"good\" template in the background, but I do not think this is a mixed colony or diff template."},
//	"2eZf1XbOe1Rk2SuxtUdy5dPjZGx": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed\n\n165 A->G nanopore sequencing error."},
//	"2eZf2SNEcpeggaZPgAwrL4b0Zpf": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed. 165 A->G is a nanopore error."},
//	"2eZf2glq5x7Hhh3MLUW1Y3yLwxl": {Confirmed: false, MixedTemplate: true, MixedColony: true, Notes: "Mutated. This appears to be a mixed colony with different templates...."},
//	"2eZf2x3v57l8AcpPnVg6rz70gB4": {Confirmed: false, MixedTemplate: true, MixedColony: false, Notes: "-189 deletion at 132 that VCF did not pick up? We got 1 read, but it appears to have changed 310 C->G, 311 C->A, 312 G->C, so I think this is a different template somehow getting read. Please note all that."},
//	"2eZf36GSMUYEAiIeh76Q3sfa4Ac": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed.\n\nNote the following: the 165 mutation here is only effecting one strand, not the other. This is how you can tell if it is a nanopore error."},
//	"2eZf3GQwjBxPBVvbxCtbU56D4xq": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "mutated. 632 G->A. Not a nanopore sequencing error. Appears to be a mixed colony of same template."},
//	"2eZf3et1n64i3Txw8tFHKMx6Bje": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed\n\n165 A->G nanopore sequencing error."},
//	"2eZf47WjgSHf6CqJjhuwU35oaRk": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed\n\n165 A->G nanopore sequencing error."},
//	"2eZf4CVBcgELd2Zp71pPFu4Bscx": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "mutated 221 G->T. not a nanopore sequencing error. Appears mixed colony, same template"},
//	"2eZf4gCcdQXgaZ0VmzyN79wXUIb": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "mutated 308 G->T. not nanopore error. mixed colony, same template."},
//	"2eZf4qG0KhgmG9ikWg9hMXaIOdK": {Confirmed: false, MixedTemplate: true, MixedColony: true, Notes: "Call it mutated, but really it is just a mixed colony AND different template."},
//	"2eZf55Bf0MAYnoLGuBYu38zEmlx": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "\"mutated\", but really just wrong template. (diff template, not mixed)"},
//	"2eZf85ohqMnwSFX8ve9c0uCd4us": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed\n\n165 A->G nanopore sequencing error."},
//	"2eZf8ZFF6H2a2vmUQVhY9VHwcYU": {Confirmed: false, MixedTemplate: true, MixedColony: true, Notes: "Mutated. Mixed colony, different template."},
//	"2eZf8VmvopeWG2VSdTKankMxcVv": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed\n\n458 A->G, 461 A->G nanopore sequencing error"},
//	"2eZf9qhEPrDOxSmqoXtQvpmAnFX": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed\n\n165 A->G nanopore sequencing error."},
//	"2eZf9zAPzNE2tEg1w9GbEE5rGsI": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "Low sequencing cover."},
//	"2eZfA1rLTlmoAzIoRTeSVDH2mA5": {Confirmed: false, MixedTemplate: true, MixedColony: true, Notes: "mutated. 818 +1G, 826 T->C, 827 C-G, 831 -1C. Mixed colony, different templates."},
//	"2eZfAcZPiq7unqIrBPdPpeyJmBo": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "mutated 275 T->C. Clean. not mixed, not diff template."},
//	"2eZfBgbAulY4SQZzQqMDKopUkoZ": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "mutated 770 +1G. Clean."},
//	"2eZfCmTtG8CrSYXIjd8bPemXJj5": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "mutated. 218 G->T, 252 A->T, 255 G->T, 265 A-T. Clean."},
//	"2eZfD77XQMVThSJ8Sz5HEVjUbUR": {Confirmed: false, MixedTemplate: true, MixedColony: true, Notes: "mutated. 176 +1C, 189 -1G, 334 G->A, 336 G->C, (more mutations as well). mixed colony, different template."},
//	"2eZfDUq0DlFSEaWe1ZspqqS5MCw": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "mutated. Recircularized vector."},
//	"2eZfDt7gCYNDh1nliDvvAQmdqc0": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed\n\n165 A->G nanopore sequencing error."},
//	"2eZfE58cWYp9CUMbCgjxPe5118d": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "mutated. 517 C->T, 806 -2TC. There are a couple correct reads, so same template, mixed colony."},
//	"2eZfETtl7MoveWCgoGPKPTK18NL": {Confirmed: false, MixedTemplate: true, MixedColony: false, Notes: "mutated. Not mixed colony, but different template."},
//	"2eZfFYa7pcWOAJb6NpPWCHBdikz": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "829 G->T mutation. Clean."},
//	"2eZfFYBxGahtJib7FAOS8byYOOi": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed.\n\n670 A->C, but is a nanopore sequencing error.\n709 C->A, but is a nanopore sequencing error.\n710 G->C, but is a nanopore sequencing error."},
//	"2eZfGbC6iIjpNn7O4qkHxsjeUZg": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "849 G->A mutation. Clean"},
//	"2eZfGVcXAJiFrGcDI6gLrBpPa5z": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "165 A->G nanopore sequencing error, but 373 G->T real mutation. Clean."},
//	"2eZfH72Z55Sjeqr1LSSbft2vVt1": {Confirmed: false, MixedTemplate: true, MixedColony: true, Notes: "mixed colony, different template. mutated."},
//	"2eZfHAeRucChcng83lf9uHUXE31": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed\n\n165 A->G nanopore sequencing error."},
//	"2eZfI6sQAdP30dUTGspxj91CxYz": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed\n\n165 A->G nanopore sequencing error."},
//	"2eZfJ10fLZiOZ22G4DlnlPbedBY": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed\n\n165 A->G nanopore sequencing error."},
//	"2eZfJ28qbEFd8oIgPuLstMsEVCy": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "Confirmed\n\n165 A->G nanopore sequencing error."},
//	"2eZfJEpvK5htDQMRZxeuPC5bXWi": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "428 -139 deletion. Mutated"},
//	"2eZfJOBWDUJ1j5hXhw9M81vx0gG": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "136 -5ATATA deletion, 165 G->A. Clean."},
//	"2eZfJtP3BqJlgoAMBVzpH1d7HYq": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "166 -1G. Clean"},
//	"2eZfKHRAAkc9DuKxFFr4Wo2GEiF": {Confirmed: false, MixedTemplate: true, MixedColony: true, Notes: "Mixed colony, different template."},
//	"2eZfKrtmuEEB8rSDhM4Aoowuwh1": {Confirmed: false, MixedTemplate: true, MixedColony: false, Notes: "different template, non-mixed colony."},
//	"2eZfLSEiGqlDxdyciBAtMeKK26X": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "255 G->A. Clean"},
//	"2eZfLkhYPzcK80dHijVuchnUpRT": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "mixed colony, same template. Indel at 195 of \"ATTTTTTTTT\"."},
//	"2eZfLyCs9DCR8KX3vI6SoY72Hvr": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "689 G->T. Mixed colony, same template."},
//	"2eZfM1JELsIG2vf7GWowdwo7CoF": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "165 A->G nanopore sequencing error. 543 -1G, mixed colony, same template"},
//	"2eZfM1oQU67fYZsKgUEO5VfbN3g": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "confirmed. 570 A->G nanopore sequencing error."},
//	"2eZfbNtTcYquLi08sMuEiWOGVc8": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "Nanopore sequencing error of 177 -1, but also 292,293->G, 308--, 325 A->C, 330 G->C,and 336--"},
//	"2eZfhwDpmEKfmVznW9Swr5Y5pmw": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "831 C->T. Mixed colony, same template."},
//	"2eZfivJl5zIg6qiblJpipmpuHwv": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "114 C->A. Mixed colony, same template."},
//	"2eZftKXHVD76Fu3vT7m2g2IGMJ0": {Confirmed: false, MixedTemplate: true, MixedColony: true, Notes: "mixed colony + mixed template."},
//	"2eZg0K3QP4PsTNsmaFzwuRS2SYU": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "254 -3. Clean."},
//	"2eZg0jyZm7pw3Skh75HkazfGuo9": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "807 -51 deletion. Clean"},
//	"2eZg2dy6hpyk5xZUvT0wee7G4Wn": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "336 G->T. Very clean."},
//	"2eZg95ju6huhLqrnojlbUanYcEe": {Confirmed: false, MixedTemplate: true, MixedColony: false, Notes: "Not mixed colony. Different template, though."},
//	"2eZgB0xuOolucCdyinmHVKsZkao": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "mutated, clean, 542 -2"},
//	"2eZgE7Y7of5R8LaqxEiMZaQU7SM": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "mixed colonies, same template. 128 C->G"},
//	"2eZgHcMVkFEKqcLlPYhpNdtLQRY": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "250 -1. Clean."},
//	"2eZgIHXbCLd26sxcMRfMe4IhFp2": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "heavily mutated, clean, mutations: 118 G->T, 122 C->T, 139 -2, 142 G->T, 157 G->T, 160 C->T, 165 C->T, 166 C->T, 178 C->T, 181 -1"},
//	"2eZgIT8bfzZNq75gKwJv1NI6sXh": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "338 C->T. Mixed colony."},
//	"2eZgJXOkepQiAN0PYAaGQzVPXS2": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "191 indel. Clean."},
//	"2eZgQoz16OkrrfuRAFHwP8gZHj0": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "mutated, clean, 418 -1"},
//	"2eZgVFjanzTPsn0ObgSHGHSZudE": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "175 G->A clean"},
//	"2eZgVgV4pD8Z3jWmy8KuE5TBHqX": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "151 to 189 is deleted. Mixed colonies."},
//	"2eZgWhPmKRg9HTadTuzU0tMGm6F": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "120 A->T, 162 T->A, 171 T->G. mixed colony."},
//	"2eZgd1flppPwFLVMP9LG8sCGwxI": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "137 -1T, 230 -2, clean"},
//	"2eZghrdswZ0ooDo577yniNLP712": {Confirmed: false, MixedTemplate: true, MixedColony: true, Notes: "different template, mixed colony."},
//	"2eZgvuuL1zYlsT8rptBdyJ8pDaT": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "Low read count"},
//	"2eZgwiGBr1gHM2Rx5cj1axOIFHU": {Confirmed: true, MixedTemplate: false, MixedColony: true, Notes: "confirmed\n\nHowever, there is a mixed colony with non-mixed template."},
//	"2eZh7sfIOijSONuvoOzZR34s9V9": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "wrong template. non-mixed colonies"},
//	"2eZhEPeswUYMiicn32dHYS2av6X": {Confirmed: false, MixedTemplate: true, MixedColony: false, Notes: "not mixed, but different template"},
//	"2eZhFPuuMyIxsKpUQ6DxKen6nhO": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "895 A->T. Clean."},
//	"2eZhLAFIMsEBsL6ms9yLnZPwQY9": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "159 -4, clean"},
//	"2eZhMryC1HRHqGx6hoKSRmfY0I6": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "225 -1. clean"},
//	"2eZhSOYqMSPng9ZALW2A0wO9nAs": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "240 -2. clean."},
//	"2eZhUR7uW1oTX5v4x8lDqC06HTD": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "140,255 indel"},
//	"2eZhgeHQlJi6wuOID5ai2vfa8P1": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "cloning issue"},
//	"2eZhhu4jVPYS28H094XHte3qnWq": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "190 G->A, clean"},
//	"2eZhjBPsRVsp3WSv0WMtlTaPsxK": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "317 -1, 384 +1G. Mixed colonies."},
//	"2eZhl5gO0ZhBw3UjOUcGTmq6Hgd": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "225 -1. clean."},
//	"2eZhoD9kuz4JgdQL9S1vovLUoL1": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "124 -4, clean"},
//	"2eZhqOo9nGmH1ZpD3xOBlSUPmRE": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "low read count."},
//	"2eZhuaczcVJFoANnavDDNRU52NE": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "897 -3, clean."},
//	"2eZhwYGEvV9WEScwtEi368nr7bN": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "146 -7 clean."},
//	"2eZhxL7aTAUoajohU9azLw2u6x1": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "confirmed (human)"},
//	"2eZhxOq0Am2wJLHLnYGmSk8OdWo": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "169 -2,clean"},
//	"2eZi14UzWjIFn6kk3sjcMLU5LTm": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "184 A->T. Clean."},
//	"2eZi1YOjA55BmPfXy45OGb0uigG": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "123 C->A, clean"},
//	"2eZi2FypPrzJZ72OKNXCDHLLFMn": {Confirmed: false, MixedTemplate: true, MixedColony: false, Notes: "different template, but not mixed colonies."},
//	"2eZi3fENGQV4H0vpSg8dxiBLwvz": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "181 A->G. clean."},
//	"2eZi7W7jihRDVm38zojgVy17Jzj": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "158 -1, 162 -2. clean."},
//	"2eZi9LxkZzWe3hsW354vhvnJg8d": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "720 G->T clean"},
//	"2eZiAG0catuntiRsPWbNKkNyKyZ": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "121 A->T, 141 G->A, 163 G->A, 167 C->T, 228 -2. Clean."},
//	"2eZiBUZcxBxQwK3O1tnGXWPro7s": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "127 -3. clean"},
//	"2eZiC7H8AihwUOS7WQMzNLhc4YT": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "183 -5. clean."},
//	"2eZiCgEKzv54UUSeMgCHgqWWZtP": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "144 -3 clean"},
//	"2eZiEVUbgCR9AdBzeyR9k8uig1t": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "153 G->A. 169 -1. Clean."},
//	"2eZiFjamRJEaPJxulCiqPL0rGM3": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "146 -11. Mixed colonies."},
//	"2eZiL1L0lhMhl8v8o9RkH79ZlCm": {Confirmed: true, MixedTemplate: true, MixedColony: true, Notes: "Confirmed, but mixed colonies / different template."},
//	"2eZiNXTHqeJfIEp6qnlxtDE0YUB": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "116 -1, clean."},
//	"2eZiPV8vG5T9CSw5EAazI9FAvT4": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "wrong template"},
//	"2eZiTwQoFV6lyZs99wFQsTU3QF3": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "584 T->A. Clean."},
//	"2eZiVTWgAOEQPKUpUWR0PSPIA8r": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "167 T->G, 176 A->G, 257 G->A. Mixed colony."},
//	"2eZiaREUkcLXmzmGk5jLpvLBNsL": {Confirmed: true, MixedTemplate: false, MixedColony: true, Notes: "confirmed, mixed colony though. Has mutations at 852 A->T and 867 T->G at minority levels."},
//	"2eZibu1fK4KB2DTvpsCaz5spsok": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "cloning issue."},
//	"2eZidNJ8HJxM5FOmR8k5A97deIW": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "219 G->C, 261 -1, clean"},
//	"2eZieUxmPlsbNNwfdWGgq1DpRWQ": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "confirmed (human)"},
//	"2eZihDEEjeQPDqWWdgWwdnZk2JB": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "112 -2A mutation. Clean."},
//	"2eZimBkKRdfjetj83ypXjwbLkJ8": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "169 -1, mixed colonies"},
//	"2eZj5l0Uadoxnlqy0lItbVB7u4d": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "different template clean non-mixed colonies"},
//	"2eZjDHhgmq7XYOgnBe4nROLOt4V": {Confirmed: false, MixedTemplate: true, MixedColony: false, Notes: "different template. non-mixed colony."},
//	"2eZjDY8dIpviNGYQU1QadTdRGuC": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "different template, but non-mixed."},
//	"2eZjFwDDFPwyxno4r8muJupMK95": {Confirmed: false, MixedTemplate: true, MixedColony: false, Notes: "different template, non-mixed colonies"},
//	"2eZjJLnsKC91WYcfwhXCev3HAwS": {Confirmed: false, MixedTemplate: true, MixedColony: false, Notes: "different template, but not mixed colonies."},
//	"2eZjRhAzuxdfJcgFK4Mrb8h711B": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "different template"},
//	"2eZjRrJ4VxhpnthQYsFnh1WLmk8": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "different template, non-mixed. Mutated."},
//	"2eZjV8QFDBc0ioBp8f925hQfV71": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "confirmed (human annotated)"},
//	"2eZjaLkVJy3VbMpUeJhwyo7zYjL": {Confirmed: false, MixedTemplate: true, MixedColony: false, Notes: "not mixed, but different template. Also low read count"},
//	"2eZjhqSfTEVwgy4b1jTp9lU6o6J": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "different template, but not mixed."},
//	"2eZjppAVZ20HsDX78wlBvxHU9aK": {Confirmed: false, MixedTemplate: false, MixedColony: true, Notes: "397 T->A, 399 C->G, 472 G->A, mixed colony, same template."},
//	"2eZjrrplQ3lZrX8pBpxvrGqwiIg": {Confirmed: false, MixedTemplate: true, MixedColony: true, Notes: "mixed template and mixed colonies."},
//	"2eZjsgdWGoY3cODgmDph2oj3jO3": {Confirmed: false, MixedTemplate: true, MixedColony: false, Notes: "different template, non-mixed colonies."},
//	"2eZjth2sn8byWkDfLjOt9Wful6N": {Confirmed: true, MixedTemplate: false, MixedColony: false, Notes: "confirmed (human)"},
//	"2eZjwltxf5hx4lTeTtx46HhXGmZ": {Confirmed: false, MixedTemplate: true, MixedColony: false, Notes: "different template, non-mixed colonies."},
//	"2eZk9yrkU6M0wRw8XunXHmhKc9B": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "low read count."},
//	"2eZkCr3v7apysiMEESUc5irN8nH": {Confirmed: false, MixedTemplate: true, MixedColony: true, Notes: "mixed colonies and mixed template."},
//	"2eZkFDM5jjwHHJG16nP7j2HWbFx": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "low read count"},
//	"2eZkPU6OBMpQGTsVfqA7kRO8bJc": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "114 point A G"},
//	"2eZkQpl6amNF5DuhGIhFlXqIjoq": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "different template. non-mixed colonies."},
//	"2eZkRQ2FBDXz4BOiSNKQcZXFdGq": {Confirmed: false, MixedTemplate: false, MixedColony: false, Notes: "different template. mutated."},
//}
//
//func readPileupFiles(dirPath string) (map[string][]Line, error) {
//	fileMap := make(map[string][]Line)
//
//	files, err := os.ReadDir(dirPath)
//	if err != nil {
//		return nil, err
//	}
//
//	for _, file := range files {
//		if filepath.Ext(file.Name()) == ".pileup" {
//			filePath := filepath.Join(dirPath, file.Name())
//			fileName := strings.TrimSuffix(file.Name(), ".pileup")
//
//			f, err := os.Open(filePath)
//			if err != nil {
//				return nil, err
//			}
//			defer f.Close()
//
//			const maxLineSize = 2 * 32 * 1024
//			parser := NewParser(f, maxLineSize)
//			var pileupReads []Line
//
//			for {
//				pileupRead, err := parser.Next()
//				if err != nil {
//					if !errors.Is(err, io.EOF) {
//						return nil, err
//					}
//					break
//				}
//				pileupReads = append(pileupReads, pileupRead)
//			}
//
//			fileMap[fileName] = pileupReads
//		}
//	}
//
//	return fileMap, nil
//}
//
//func TestSequencingResults(t *testing.T) {
//	dirPath := "./data/sequencing"
//	fileMap, err := readPileupFiles(dirPath)
//	if err != nil {
//		t.Fatalf("Failed to read pileup files: %s", err)
//	}
//
//	for sequenceName, manualResult := range sequencingResults {
//		if manualResult.Confirmed {
//			computedResult, err := GetSequencingResult(fileMap[sequenceName], "GGTCTC", "GAGACC", 0.4)
//			if err != nil {
//				t.Errorf("Got err on GetSequencingResult on %s: %s", sequenceName, err)
//			}
//			if !computedResult.Confirmed {
//				t.Errorf("Failed to get confirmed result on %s. Got: %v", sequenceName, computedResult)
//			}
//		}
//	}
//}
