<p align="center">
  <img src="media/lemc-readme-logo.png" alt="Let'em Cook! Logo" width="400"/>
</p>

# Let'em Cook! 🔥 (LEMC)

## What problem is `LEMC` solving?

Have you ever found yourself thinking 
- _"I wish I could clone myself I'm oversubscribed!"_
- _"We have such a small team of coders & hackers, how could I force multiply them?"_
- _"I wonder if I could even leverage my boss or my customer to act as an extra set of hands?"_ 
- _"If only a button existed that could be clicked that would just do-the-thing..."_ 

Then `LEMC` is made for you!

## The `LEMC` thesis

In my experience when working on a team of developers or hackers, they often need to perform deterministic tasks, on non-deterministic schedules, and then communicate the results in a streamlined fashion.

Far too often these tasks fall under the domain of DevOps and so something like `Rundeck`, `Jenkins`, `GitHub Actions`, `GitLab CI/CD`, `Ansible AWX/Tower`, `Puppet Bolt`, `SaltStack`, `Chef Automate`, `Argo Workflows`, or `Apache Airflow` is implemented to help manage all-the-things. Unfortunately, these pieces of software can take a lot of support and tend to require a DevOps team with specialized knowledge as they are built for the enterprise market. This results in DevOps not acting as an extension of the team but rather its very own department. This causes a lot of friction, creating organizational drag, thus limiting the speed at which the team can ship.  

`LEMC` is built with the belief that in the age of vibe-coding, developers will out-pace their competition when they own their operations. 

The ultimate goal being that `LEMC` would help your organization **"Ops their Devs."**

## How does `LEMC` help?

`LEMC` works to free siloed code or business logic. It's the type of code that gets core business work done but tends to sit on someone's computer running under their desk or may need that "special engineer" around to run it manually. And when said engineer is out-of-office, suddenly the organization is screwed.

`LEMC` allows anyone on your team to take their siloed code or lone-wolf scripts, wrap them in a container, and quickly empower their team to get visual results streamed to the browser right from inside the container at the click of a button. It does this with a few special verbs and the most used programming functions of all time, `print` or `echo`. 

`LEMC` is built in anticipation of AI-assisted code generation which helps fast moving teams build and innovate quickly. Forsaking many modern and GUI-heavy solutions, `LEMC` is a **language first solution**. This is perfect for LLM vibe-coding sessions. 
## Technology Stack

*   **Backend:** Go (Golang 1.23.0)
*   **Web Framework:** [Echo](https://echo.labstack.com/)
*   **Templating:** [Templ](https://templ.guide/)
*   **Frontend Interaction:** [HTMX](https://htmx.org/)
*   **Database:** SQLite
*   **Database Interaction:** [sqlx](https://github.com/jmoiron/sqlx), [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
*   **Database Migrations:** [pressly/goose](https://github.com/pressly/goose)
*   **Tailwind CSS** [tailwindcss](https://tailwindcss.com/)
*   **Scheduling:** [go-quartz](https://github.com/reugn/go-quartz)
*   **Container Interaction:** Docker SDK
*   **Realtime:** WebSockets ([gorilla/websocket](https://github.com/gorilla/websocket))
*   **Containerization:** Docker, Docker Compose

## LEMC Core Development Requirements

To contribute to or develop `LEMC`, you'll need:

*   **Go:** Version 1.23.0 (check `.go-version` or `.goenv`).
*   **Docker & Docker Compose:** For running the application containerized and executing recipes.
*   **Access to a Docker Daemon Socket:** Required for `LEMC`'s core container interaction features (default `unix:///var/run/docker.sock`).
*   **[air](https://github.com/cosmtrek/air):** Recommended for live reloading during development (`go install github.com/cosmtrek/air@latest`).
*   **[templ](https://templ.guide/):** Required for compiling `.templ` files into Go code (`go install github.com/a-h/templ/cmd/templ@latest`).
*   **Git:** For version control.

## Special Thanks

I just wanted to say thanks to [Ed Skoudis](https://x.com/edskoudis) and the [CounterHack.com](https://www.counterhack.com/) team for always encouraging me to push myself! To allow for personal time and space to educate and innovate.

Ed you are truly a wonderful man and I'm thankful you are in my life.

<p align="left">
  <img src="media/counter-hack-white.png" alt="CounterHack Logo" width="125"/>
</p>



## Copyright

Let'em Cook!, LEMC, and all associated logos and assets are the copyright of Jared Folkins. 
Copyright © 2025 Jared Folkins. 
All rights reserved.