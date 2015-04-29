package main

import (
	"time"
)

const welcomeSubject = "Welcome to Lavaboom"

const welcomeTemplate = `<p>{{.FirstName}},</p>

<p>I'm delighted to have you on board.</p>

<p>Lavaboom is currently in beta, a testing period for the service. Over the coming
weeks we’ll be making changes based on your feedback and rolling out new
features.</p>

<p>Drop us a line anytime by emailing the team (see the contacts tab), and follow
us on Twitter at <a href="https://twitter.com/LavaboomHQ">@lavaboomhq</a> for service updates.</p>

<p>The team will have sent you an email by now that will get you started.</p>

<p>Welcome on board.</p>

<p>Felix Müller-Irion
Lavaboom Founder</p>`

const startedSubject = "Getting started with Lavaboom"

const startedTemplate = `<p>Hey {{.FirstName}},</p>

<p>I'm Bill, and this is a quick message to get you started.</p>

<p>Below are some handy links that will help you use Lavaboom:</p>

<p>1. Lavaboom makes encryption easy. <a href="https://support.lavaboom.com">Find out how to send an encrypted email.</a></p>

<p>2. Attachments are not yet sent encrypted, we’ll let you know when this is
available. Attachments are also limited to 5mb, for larger files we suggest
using <a href="https://spideroak.com">spideroak.com</a>.</p>

<p>3. You can get in touch with Lavaboom staff anytime, head to the Contacts tab
and you’ll find our email addresses. <a href="https://mail.lavaboom.com/contacts">Go to contacts.</a></p>

<p>4. For questions, walkthroughs and support head to <a href="https://support.lavaboom.com">support.lavaboom.com</a>.</p>

<p>A note from the security team will be arriving shortly with additional
information.</p>

<p>Do you have any questions? Hit ‘reply’ and send your first secure email -
I'm happy to help.</p>

<p>Great to have you on board,</p>

<p>Bill Franklin
Lavaboom</p>`

const securitySubject = "Important security info"

const securityTemplate = `<p>Hi {{.FirstName}},</p>

<p>This is Andrei from the Lavaboom Security Team.</p>

<p>Lavaboom is built to be easy to use and remove the email provider as a threat
vector. So now the weakest link in the security chain is you and your computer.
Below are some basic pointers:</p>

<p>1. Never share your Private Key with anyone (not even us). <a href="https://support.lavaboom.com">Find out what a
Private Key is.</a></p>

<p>2. There are some things we can’t encrypt. <a href="https://support.lavaboom.com">Find out what we don’t encrypt.</a></p>

<p>3. Lavaboom is not suitable for protecting you from the NSA. However if you are
a journalist using Gmail this is a titanic step up in your security. <a href="https://support.lavaboom.com">Read more
about Lavaboom’s Threat Model.</a></p>

<p>4. If you believe you are a direct target of a government or private organisation please email
<a href="mailto:security@lavaboom.com">security@lavaboom.com</a> from this email address.</p>

<p>If you have questions about this information or want to learn more about how
Lavaboom protects you, simply reply to this email.</p>

<p>Best wishes,</p>

<p>Andrei Simionescu
Lavaboom Security Team</p>`

const whatsupSubject = "How's it going?"

const whatsupTemplate = `<p>Hey {{.FirstName}},</p>

<p>We hope you have been enjoying Lavaboom, we’re just checking in to see how
you’re getting along - how does it feel sending secure emails?</p>

<p>We’re so excited that you’re diving into Lavaboom! If you notice something
strange, want to talk encryption or just fancy a chat we are online 24/7.</p>

<p>You can find status updates on <a href="http://twitter.com/lavaboomhq">our Twitter</a>, <a href="http://facebook.com/lavaboomhq">our Facebook page</a> or email us at
<a href="mailto:help@lavaboom.com">help@lavaboom.com</a>. Find additional information on <a href="http://support.lavaboom.com">our support pages.</a></p>

<p>This message is in the >0.1% of the Internet that the NSA can’t access, so
speak freely.</p>

<p>Looking forward to hearing from you,</p>

<p>The Lavaboom Support Team</p>`

var (
	welcomeDelay  = time.Second * 5
	startedDelay  = time.Second * 30
	securityDelay = time.Minute * 3
	whatsupDelay  = time.Minute * 30
)
