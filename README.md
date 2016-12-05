Logind in Ubuntu 16.04 seems to randomly forget to actually close ssh sessions.
As we run a fair number of them per day via automatism we'll want to clean
up every once in a while.

https://phabricator.kde.org/T4690

This is fairly straight forward cleanup, could be written in any language but
given Go is easy to staticly link we'll use it here to avoid problems with
third party dbus libs potentially not being available at runtime.

# Building

- set GOPATH in environment
- install go
- go get -u github.com/mattn/gom # dependency helper
- gom -production install # install dependencies
- gom build # build

# Deploy

to deploy updates to KDE (you need access to neonarchives user)

- `gom build`
- `ruby ./deploy.rb`

the deploy.rb helper will rsync the built binary and twiddle systemd.
