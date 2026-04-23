Name:           qtop
Version:        1.0.0
Release:        1%{?dist}
Summary:        A modern, responsive TUI system monitor written in Go

License:        MIT
URL:            https://github.com/QurieGLord/QTop
BuildRequires:  golang
Requires:       glibc

%description
A modern, responsive TUI system monitor written in Go.
A clean, beautiful, and feature-rich TUI system monitor like btop.
Includes monitoring for CPU, GPU, Memory, ZRAM, Swaps and Processes.

%install
mkdir -p %{buildroot}/usr/bin
install -p -m 755 qtop %{buildroot}/usr/bin/qtop

%files
/usr/bin/qtop
