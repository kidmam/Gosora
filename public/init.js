'use strict';

var me = {};
var phraseBox = {};
if(tmplInits===undefined) var tmplInits = {};
var tmplPhrases = []; // [key] array of phrases indexed by order of use
var hooks = { // Shorten this list by binding the hooks just in time?
	"pre_iffe": [],
	"pre_init": [],
	"start_init": [],
	"end_init": [],
	"after_phrases":[],
	"after_add_alert":[],
	"after_update_alert_list":[],
	"open_edit":[],
	"close_edit":[],
	"edit_item_pre_bind":[],
	"analytics_loaded":[],
};
var ranInitHooks = {}

function runHook(name, ...args) {
	if(!(name in hooks)) {
		console.log("Couldn't find hook '" + name + "'");
		return;
	}
	console.log("Running hook '"+name+"'");

	let hook = hooks[name];
	for (const index in hook) {
		hook[index](...args);
	}
}

function addHook(name, callback) {
	hooks[name].push(callback);
}

// InitHooks are slightly special, as if they are run, then any adds after the initial run will run immediately, this is to deal with the async nature of script loads
function runInitHook(name, ...args) {
	runHook(name,...args);
	ranInitHooks[name] = true;
}

function addInitHook(name, callback) {
	addHook(name, callback);
	if(name in ranInitHooks) {
		callback();
	}
}

// Temporary hack for templates
function len(item) {
	return item.length;
}

function asyncGetScript(source) {
	return new Promise((resolve, reject) => {
		let script = document.createElement('script');
		script.async = true;

		const onloadHandler = (e, isAbort) => {
			if (isAbort || !script.readyState || /loaded|complete/.test(script.readyState)) {
				script.onload = null;
				script.onreadystatechange = null;
				script = undefined;

				isAbort ? reject(e) : resolve();
			}
		}

		script.onerror = (e) => {
			reject(e);
		};
		script.onload = onloadHandler;
		script.onreadystatechange = onloadHandler;
		script.src = source;

		const prior = document.getElementsByTagName('script')[0];
		prior.parentNode.insertBefore(script, prior);
	});
}

function notifyOnScript(source) {
	source = "/static/"+source;
	return new Promise((resolve, reject) => {
		let ss = source.replace("/static/","");
		try {
			let ssp = ss.charAt(0).toUpperCase() + ss.slice(1)
			console.log("ssp:",ssp)
			if(window[ssp]) {
				resolve();
				return;
			}
		} catch(e) {}
		
		console.log("source:",source)
		let script = document.querySelectorAll('[src^="'+source+'"]')[0];
		console.log("script:",script);
		if(script===undefined) {
			reject("no script found");
			return;
		}

		const onloadHandler = (e) => {
			script.onload = null;
			script.onreadystatechange = null;
			resolve();
		}

		script.onerror = (e) => {
			reject(e);
		};
		script.onload = onloadHandler;
		script.onreadystatechange = onloadHandler;
	});
}

function notifyOnScriptW(name, complete, success) {
	notifyOnScript(name)
		.then(() => {
			console.log("Loaded " +name+".js");
			complete();
			if(success!==undefined) success();
		}).catch((e) => {
			console.log("Unable to get script name '"+name+"'");
			console.log("e: ", e);
			console.trace();
			complete(e);
		});
}

// TODO: Send data at load time so we don't have to rely on a fallback template here
function loadScript(name, callback,fail) {
	let fname = name;
	let value = "; " + document.cookie;
 	let parts = value.split("; current_theme=");
 	if (parts.length == 2) fname += "_"+ parts.pop().split(";").shift();
	
	let url = "/static/"+fname+".js"
	let iurl = "/static/"+name+".js"
	asyncGetScript(url)
		.then(callback)
		.catch((e) => {
			console.log("Unable to get script '"+url+"'");
			if(fname!=name) {
				asyncGetScript(iurl)
					.then(callback)
					.catch((e) => {
						console.log("Unable to get script '"+iurl+"'");
						console.log("e: ", e);
						console.trace();
					});
			}
			console.log("e: ", e);
			console.trace();
			fail(e);
		});
}

/*
function loadTmpl(name,callback) {
	let url = "/static/"+name
	let worker = new Worker(url);
}
*/

function DoNothingButPassBack(item) {
	return item;
}

function RelativeTime(date) {
	return date;
}

function initPhrases() {
	console.log("in initPhrases")
	console.log("tmlInits:",tmplInits)
	fetchPhrases("status,topic_list,alerts,paginator,analytics")
}

function fetchPhrases(plist) {
	fetch("/api/phrases/?query="+plist)
		.then((resp) => resp.json())
		.then((data) => {
			console.log("loaded phrase endpoint data");
			console.log("data:",data);
			Object.keys(tmplInits).forEach((key) => {
				let phrases = [];
				let tmplInit = tmplInits[key];
				for(let phraseName of tmplInit) phrases.push(data[phraseName]);
				console.log("Adding phrases");
				console.log("key:",key);
				console.log("phrases:",phrases);
				tmplPhrases[key] = phrases;
			});

			let prefixes = {};
			Object.keys(data).forEach((key) => {
				let prefix = key.split(".")[0];
				if(prefixes[prefix]===undefined) prefixes[prefix] = {};
				prefixes[prefix][key] = data[key];
			});
			Object.keys(prefixes).forEach((prefix) => {
				console.log("adding phrase prefix '"+prefix+"' to box");
				phraseBox[prefix] = prefixes[prefix];
			});

			runInitHook("after_phrases");
		});
}

(() => {
	runInitHook("pre_iife");
	let toLoad = 2;
	// TODO: Shunt this into loggedIn if there aren't any search and filter widgets?
	notifyOnScriptW("template_topics_topic", () => {
		toLoad--;
		if(toLoad===0) initPhrases();
		if(!Template_topics_topic) throw("template function not found");
	});
	notifyOnScriptW("template_paginator", () => {
		toLoad--;
		if(toLoad===0) initPhrases();
		if(!Template_paginator) throw("template function not found");
	});

	let loggedIn = document.head.querySelector("[property='x-loggedin']").content;
	if(loggedIn=="true") {
		fetch("/api/me/")
		.then((resp) => resp.json())
		.then((data) => {
			console.log("loaded me endpoint data");
			console.log("data:",data);
			me = data;
			runInitHook("pre_init");
		});
	} else {
		me = {User:{ID:0,Session:""},Site:{"MaxRequestSize":0}};
		runInitHook("pre_init");
	}
})();