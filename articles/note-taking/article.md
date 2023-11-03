<sup>Photo by <a href="https://unsplash.com/@dtravisphd?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText">David Travis</a> on <a href="https://unsplash.com/photos/5bYxXawHOQg?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText">Unsplash</a></sup>

## Introduction

Taking notes is like creating a swap partition for your brain. This extension of your random-access memory enables you to offload lower-priority information. This low-effort habit can decrease your mental load and increase productivity.

Learn to take notes effectively by emphasizing simplicity and integrating the habit into your existing workflow. You will realize the benefits in your day-to-day work all while creating a record of your knowledge and achievements.


## Not just for school

I'm sure we are all familiar with taking notes from school. In my experience, this was a tedious requirement with only a few actual benefits:
1. The information prioritized by the teacher in class is more likely to be the subject of a test question, so it's good to keep track of these topics
2. Writing things down can help commit it to memory
3. Looking busy in class will reduce your chances of getting called on by the teacher

Some of these motivations are irrelevant in a work environment, but note-taking remains a valuable skill that can serve the following purposes:
- Creates a reference that is specific to your work experience. In a career where you are constantly learning new things, this is a huge asset
- Helps reduce the mental load of context-switching when going to meetings, switching between codebases, and getting interrupted by urgent bugs
- Helps you organize and manage the flood of new information when starting a new job

To be more specific, taking notes allows you to keep track of:
- Links for hard-to-find documentation or guides that you might frequently reference
- Action items from a meeting
- Complicated commands that you might need to use again
- Steps of a complex or tedious process
- Interesting code snippets
- TODO items or small bugs that aren't significant enough to go into an issue tracker
- Concrete achievements and topics for annual performance reviews
- Different solutions you have attempted when solving a complex problem

In addition to recording information, your notes become a place where you can let your thoughts flow. You can have a one-sided conversation in your notes, similar to [Rubber duck debugging](https://en.wikipedia.org/wiki/Rubber_duck_debugging), but with the ability to re-read your thoughts and ideas.


## Organizing

I encourage you to get started taking notes today. Don’t worry about organization until it becomes a necessity. This will allow you to form the habit of low effort note-taking in a way that feels natural to you. Later, you can begin to introduce more complexity to the process without interrupting your flow.

Once you get serious about taking notes, you will need a way to organize them. My preferred method is to organize by time. I create a filesystem structure around the year and month with filenames based on the week, so it ends up looking like this:
```shell
2023
└── 07Jul
    ├── Week_of_the_17th.md
    ├── Week_of_the_24th.md
    └── Week_of_the_31st.md
```

One of these files might look like:
```md
# Week of the 31st (July 2023)

### Monday 31


### Tuesday 01 (August 2023)


### Wednesday 02 (August 2023)


### Thursday 03 (August 2023)


### Friday 04 (August 2023)

```

There are numerous benefits to organizing by date rather than topic:
- You don't need to find the correct place for every topic, just put it in this week's note; as long as you use consistent terminology, you can later retrieve notes relevant to a particular topic by searching
- You can easily review what you were working on last week when you come back from the weekend
- You can find concrete examples of what you achieved in a time period for annual/quarterly reviews (or new job interviews)
- It reminds you what your thought process was when you created a specific PR or commit
- The directory structure creates a place to save ad hoc files that you use for the week, like log files for bugs and drafts of documentation

The most important consideration when organizing and creating notes is simplicity. The purpose of the notes is to improve your productivity and create a smooth process for yourself. In order to further streamline my process, I created a small CLI program that can be used to automate the building of this file structure and templating out a new week's note: [`gnotes`](https://github.com/calvinmclean/gnotes).

Every Monday, I start my day by running the `gnotes` command. It will intelligently create the directory structure for the year and month, then create the base note file for this week. Even if you run the command on a day other than Monday, it will start the week with Monday to keep things consistent.

Writing notes in a Markdown format is also very beneficial because it allows you to have nice formatting, but still keeps a simple plaintext file that is easily searchable with `grep` or other text search tools.

## Conclusion

Now that you are familiar with the benefits of note-taking, I encourage you to give it a try. Just remember to keep it simple and find what works for you!