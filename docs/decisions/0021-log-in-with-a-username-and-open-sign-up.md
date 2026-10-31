# 0021. Log in with a username, open sign-up

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

Milestone 2 adds accounts. We need to decide what a person types to log in, and who is allowed to create an account.

## Decision

An account has a unique **username** and a password. No email for now. **Anyone can sign up**: no invites yet.

## Why

- A username needs no extra system. An email address must be checked (we send a link, the person clicks it), which needs an email sender and, for development, a fake mail server (Mailpit).
- Open sign-up is the smallest complete flow. Invites add their own tokens, expiry and abuse rules on top of it.
- The users are the owner and friends, and Reach runs on one computer for now, so open sign-up is not a real risk yet.

## Alternatives considered

- **Email and password:** the usual choice, and it allows password reset by email, but it needs email verification first.
- **Username and email:** the most complete, but two things to build at once.
- **Invite-only sign-up:** safer for a public server, and listed as an idea for this milestone in the brief. It can be added on top later.

## Consequences

- Without email, a forgotten password can't be reset yet. That is fine on a development machine.
- Usernames must be unique. We must decide simple rules for them (allowed characters, length, and whether "Sadra" and "sadra" are the same name) when we build sign-up.
- Before Reach runs on a real server (milestone 10), revisit invites, email, or both.
