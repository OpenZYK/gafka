%define name kconsole
%define version 0.2.6
%define release 5
%define path usr
%define group Development/Tools
%define __os_install_post %{nil}

Summary:    kafka console for programmers
Name:       %{name}
Version:    %{version}
Release:    %{release}
Group:      %{group}
Packager:   Funky Gao <funky.gao@gmail.com>
License:    Apache
BuildRoot:  %{_tmppath}/%{name}-%{version}-%{release}
Prefix:     /usr
AutoReqProv: no

%description 
kafka console for programmers powered by golang.

%prep
mkdir -p $RPM_BUILD_DIR/%{name}-%{version}-%{release}
cd $RPM_BUILD_DIR/%{name}-%{version}-%{release}
git clone https://github.com/funkygao/gafka

%build
cd $RPM_BUILD_DIR/%{name}-%{version}-%{release}/gafka
./build.sh -t kconsole

%install
export DONT_STRIP=1
rm -rf $RPM_BUILD_ROOT
cd $RPM_BUILD_DIR/%{name}-%{version}-%{release}/gafka
mkdir -p $RPM_BUILD_ROOT/%{path}/bin
install cmd/kconsole/kconsole $RPM_BUILD_ROOT/%{path}/bin

%files
/%{path}/bin/kconsole