var guid = 'dd7f8228-2f45-4b63-9bdc-87989e693204';		//guid пользователя
var baseUrl = '';	//путь к api
var tokens = [];	//массив токенов

// запросить токены
async function getTokens(){
	let url = baseUrl + 'get-tokens/' + guid;
	let response = await sendAjax(url);
	if ( !response.Access_Token || !response.Refresh_Token ){
		alert("Ошибка запроса: отсутствуют параметры");
		console.log(response);
		throw new Error("Ошибка запроса: отсутствуют параметры");
	}


	let id = tokens.length;
	tokens.push({
		id: id,
		access: response.Access_Token,
		refresh: response.Refresh_Token,
	});



	// создание новых кнопок
	let btn = document.createElement("button");
	btn.innerText = "Обновить токен " + (id + 1);
	btn.setAttribute('data-id', id);
	btn.setAttribute('onclick', 'refreshTokens(this)');

	let btn2 = document.createElement("button");
	btn2.innerText = "Удалить токен " + (id + 1);
	btn2.setAttribute('data-id', id);
	btn2.setAttribute('onclick', 'deleteTokens(this)');

	document.querySelector("#form2").appendChild(btn);
	document.querySelector("#form2").appendChild(btn2);


	alert("Токены запрошены");
}



// удалить все токены
async function deleteAllTokens(){
	let url = baseUrl + 'delete-all-tokens/' + guid;
	let response = await sendAjax(url);

	//очистка токенов
	tokens = [];
	// удаление кнопок для токенов
	document.querySelector("#form2").innerHTML = '';


	alert("Токены удалены");
}



// удалить одну пару токенов
async function deleteTokens(elem){
	let id = elem.getAttribute('data-id');
	let refreshToken = tokens[id]['refresh'];


	let url = baseUrl + 'delete-token/' + guid + '/' + refreshToken;
	let response = await sendAjax(url);


	delete tokens[id];

	document.querySelector('button[data-id="' + id + '"]').remove();
	document.querySelector('button[data-id="' + id + '"]').remove();

	alert("Пара токенов удалена");
}



// обновить одну пару токенов
async function refreshTokens(elem){
	let id = elem.getAttribute('data-id');
	let refreshToken = tokens[id]['refresh'];


	let url = baseUrl + 'refresh-tokens/' + guid + '/' + refreshToken;
	let response = await sendAjax(url);
	if ( !response.Access_Token || !response.Refresh_Token ){
		alert("Ошибка запроса: отсутствуют параметры");
		console.log(response);
		throw new Error("Ошибка запроса: отсутствуют параметры");
	}


	tokens[id] = {
		id: id,
		access: response.Access_Token,
		refresh: response.Refresh_Token,
	};


	alert("Пара токенов обновлена");
}



// обёртка для ajax-запросов
async function sendAjax(url, options = null){
	let response = await fetch(url, options);
	if ( !response.ok ) {
		alert("Ошибка запроса: " + response.status);
		throw new Error("Ошибка запроса: " + response.status);
	}
	let json = await response.json();
	if ( json.error && json.error != '' ){
		alert("Ошибка запроса: " + json.error);
		throw new Error("Ошибка запроса: " + json.error);
	}
	return json;
}
