# Introduction

Dgen is a simple encoder that lets you define your own character set. The character set is subdivided into character groups which are optionally required. This makes it ideal for deterministically generating a string which must contain required characters.

# How it works

The input is first hashed (SHA256) to create a seed. Dgen implements a xorshift32 based pseudo randon number generator (PRNG) which takes the seed as input. The encoder then collects characters from the required character groups based on the PRNG. The remaining characters are then chosen from the entire character set, also PRNG based. And then lastly, the characters are shuffled using Fisher–Yates.

# Why?

The intention was to create a very simple encoding algorithm that was not tied down to any particualar language (hence the implementation of a simple PRNG) and was portable accross systems.

