import{S as q,i as B,s as F,e as v,b as j,f as h,g as y,h as m,N as C,O as J,k as O,P as Y,n as z,t as N,a as P,o as w,w as E,r as S,u as A,x as R,M as G,c as H,m as L,d as Q}from"./index.662e825a.js";function D(o,e,l){const s=o.slice();return s[6]=e[l],s}function K(o,e,l){const s=o.slice();return s[6]=e[l],s}function M(o,e){let l,s,g=e[6].title+"",r,i,n,k;function c(){return e[5](e[6])}return{key:o,first:null,c(){l=v("button"),s=v("div"),r=E(g),i=j(),h(s,"class","txt"),h(l,"class","tab-item svelte-1maocj6"),S(l,"active",e[1]===e[6].language),this.first=l},m(u,_){y(u,l,_),m(l,s),m(s,r),m(l,i),n||(k=A(l,"click",c),n=!0)},p(u,_){e=u,_&4&&g!==(g=e[6].title+"")&&R(r,g),_&6&&S(l,"active",e[1]===e[6].language)},d(u){u&&w(l),n=!1,k()}}}function T(o,e){let l,s,g,r,i,n,k=e[6].title+"",c,u,_,p,f;return s=new G({props:{language:e[6].language,content:e[6].content}}),{key:o,first:null,c(){l=v("div"),H(s.$$.fragment),g=j(),r=v("div"),i=v("em"),n=v("a"),c=E(k),u=E(" SDK"),p=j(),h(n,"href",_=e[6].url),h(n,"target","_blank"),h(n,"rel","noopener noreferrer"),h(i,"class","txt-sm txt-hint"),h(r,"class","txt-right"),h(l,"class","tab-item svelte-1maocj6"),S(l,"active",e[1]===e[6].language),this.first=l},m(b,t){y(b,l,t),L(s,l,null),m(l,g),m(l,r),m(r,i),m(i,n),m(n,c),m(n,u),m(l,p),f=!0},p(b,t){e=b;const a={};t&4&&(a.language=e[6].language),t&4&&(a.content=e[6].content),s.$set(a),(!f||t&4)&&k!==(k=e[6].title+"")&&R(c,k),(!f||t&4&&_!==(_=e[6].url))&&h(n,"href",_),(!f||t&6)&&S(l,"active",e[1]===e[6].language)},i(b){f||(N(s.$$.fragment,b),f=!0)},o(b){P(s.$$.fragment,b),f=!1},d(b){b&&w(l),Q(s)}}}function U(o){let e,l,s=[],g=new Map,r,i,n=[],k=new Map,c,u,_=o[2];const p=t=>t[6].language;for(let t=0;t<_.length;t+=1){let a=K(o,_,t),d=p(a);g.set(d,s[t]=M(d,a))}let f=o[2];const b=t=>t[6].language;for(let t=0;t<f.length;t+=1){let a=D(o,f,t),d=b(a);k.set(d,n[t]=T(d,a))}return{c(){e=v("div"),l=v("div");for(let t=0;t<s.length;t+=1)s[t].c();r=j(),i=v("div");for(let t=0;t<n.length;t+=1)n[t].c();h(l,"class","tabs-header compact left"),h(i,"class","tabs-content"),h(e,"class",c="tabs sdk-tabs "+o[0]+" svelte-1maocj6")},m(t,a){y(t,e,a),m(e,l);for(let d=0;d<s.length;d+=1)s[d].m(l,null);m(e,r),m(e,i);for(let d=0;d<n.length;d+=1)n[d].m(i,null);u=!0},p(t,[a]){a&6&&(_=t[2],s=C(s,a,p,1,t,_,g,l,J,M,null,K)),a&6&&(f=t[2],O(),n=C(n,a,b,1,t,f,k,i,Y,T,null,D),z()),(!u||a&1&&c!==(c="tabs sdk-tabs "+t[0]+" svelte-1maocj6"))&&h(e,"class",c)},i(t){if(!u){for(let a=0;a<f.length;a+=1)N(n[a]);u=!0}},o(t){for(let a=0;a<n.length;a+=1)P(n[a]);u=!1},d(t){t&&w(e);for(let a=0;a<s.length;a+=1)s[a].d();for(let a=0;a<n.length;a+=1)n[a].d()}}}const I="pb_sdk_preference";function V(o,e,l){let s,{class:g="m-b-base"}=e,{js:r=""}=e,{dart:i=""}=e,n=localStorage.getItem(I)||"javascript";const k=c=>l(1,n=c.language);return o.$$set=c=>{"class"in c&&l(0,g=c.class),"js"in c&&l(3,r=c.js),"dart"in c&&l(4,i=c.dart)},o.$$.update=()=>{o.$$.dirty&2&&n&&localStorage.setItem(I,n),o.$$.dirty&24&&l(2,s=[{title:"JavaScript",language:"javascript",content:r,url:"https://github.com/pocketbase/js-sdk"},{title:"Dart",language:"dart",content:i,url:"https://github.com/pocketbase/dart-sdk"}])},[g,n,s,r,i,k]}class X extends q{constructor(e){super(),B(this,e,V,U,F,{class:0,js:3,dart:4})}}export{X as S};
