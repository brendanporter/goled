color = {}
color.R = 0;
color.G = 0;
color.B = 0;
color.A = 0;

fill = false;

drawmode = false;
canvasSerial = 0;

$(document).ready(function(){
	init();
});


function init() {


	$('.pixel').on('mousedown', function(event) {
		drawmode = true;
		if(event.button == 0){
			event.target.onclick.apply(event);
		}
		event.preventDefault();
	});

	$('.pixel').on('contextmenu', function(event){
		event.preventDefault();
		dropper = $(event.target).css('background-color');
		$("#color").spectrum('set', dropper);
		setColor();
	});

	$(document).on('mouseup', function(event) {
	    drawmode = false;
	});

	$.fn.spectrum.load = false;

	$("#color").spectrum({
	    change: function(color){
	    	console.log(color);
	    	setColor();
	    },
	});

	refreshDisplayFromServer();

	setInterval(refreshDisplayFromServer, 2000);

	getImages()
	getAnimations()

	

}

function setColor(){
	p = $('#color').spectrum('get').toRgb();
	color.R = p.r;
	color.G = p.g;
	color.B = p.b;
	color.A = 255;
}

function hoverPixel(x,y){
	if(drawmode) {
		setPixel(x,y);
	}
}

function saveCanvasAsImage() {

	var name = prompt('Please name the image:')
	if (name === "") {
	    // user pressed OK, but the input field was empty
	    return false;
	} else if (name) {
	    // user typed something and hit OK
	} else {
	    // user hit cancel
	    return false;
	}

	if(name === ""){
		return false;
	}

	$.ajax({
	url: "/api?action=saveCanvasAsImage",
	type: 'post',
	dataType: 'json',
	data: {name: name},
	beforeSend: function(){
	},
	success: function(json){
		getImages()
	}
	});
}

function playAnimation(name) {
	
	var loops = prompt('How many loops?')
	if (loops === "") {
	    // user pressed OK, but the input field was empty
	    loops = 3;
	} else if (loops) {
	    // user typed something and hit OK
	} else {
	    // user hit cancel
	    return false;
	}


	$.ajax({
	url: "/api?action=playAnimation",
	type: 'post',
	dataType: 'json',
	data: {name: name, loops: loops},
	beforeSend: function(){
	},
	success: function(json){
		clearDisplay();
	}
	});
}

function newAnimation() {
	var name = prompt('Please name the animation:')
	if (name === "") {
	    // user pressed OK, but the input field was empty
	    return false;
	} else if (name) {
	    // user typed something and hit OK
	} else {
	    // user hit cancel
	    return false;
	}

	if(name === ""){
		return false;
	}

	$.ajax({
	url: "/api?action=saveNewAnimation",
	type: 'post',
	dataType: 'json',
	data: {name: name},
	beforeSend: function(){
	},
	success: function(json){
		getAnimations();
	}
	});
}

function deleteAnimationFrames(name) {
	frames = [];
	$('#animations .card-header b:contains("'+name+'")').filter(function(){return $(this).text() === name;}).parent().parent().find('input:checked').each(function(i,e){
		frames.push($(e).data('frame'));
	});

	$.ajax({
	url: "/api?action=deleteAnimationFrames",
	type: 'post',
	data: {name: name, frames: frames},
	beforeSend: function(){
	},
	success: function(){
		getAnimations();
	}
	});
}

function saveFrameToAnimation(name){
	$.ajax({
	url: "/api?action=saveFrameToAnimation",
	type: 'post',
	dataType: 'json',
	data: {name: name},
	beforeSend: function(){
	},
	success: function(json){
		getAnimations();
	}
	});
}


function getFramesForAnimation(name) {
	$.ajax({
		url: "/api?action=getAnimations",
		type: 'post',
		dataType: 'html',
		data: {canvasSerial: canvasSerial},
		beforeSend: function(){
			$('#animations').html('')
		},
		success: function(html){

		}
	});
}


function getAnimations() {
	$.ajax({
	url: "/api?action=getAnimations",
	type: 'post',
	dataType: 'html',
	data: {canvasSerial: canvasSerial},
	beforeSend: function(){
		$('#animations').html('')
	},
	success: function(html){
		$('#animations').html(html)
		$('.sortable').sortable();
		//$('.sortable').disableSelection();
		$('.sortable').on('sortstop', function(event,ui){
			name = $(event.target).data('animation');
			frames = $(event.target).sortable("serialize");
			$.ajax({
				url: "/api?action=rearrangedAnimationFrames&" + frames,
				type: 'post',
				dataType: 'html',
				data: {name: name},
				beforeSend: function(){
				},
				success: function(html){
					console.log('Retrieved new frames');
					$(event.target).html(html);
				}
			});
		});
	}
	});
}

function deleteAnimation(name) {

	if(!confirm("Delete image '" + name + "'?")) {
		return false;	
	}

	$.ajax({
	url: "/api?action=deleteAnimation",
	type: 'post',
	dataType: 'json',
	data: {name: name, canvasSerial: canvasSerial},
	beforeSend: function(){
	},
	success: function(){
		getAnimations()
	}
	});
}

function deleteImage(name) {

	if(!confirm("Delete image '" + name + "'?")) {
		return false;	
	}

	$.ajax({
	url: "/api?action=deleteImage",
	type: 'post',
	dataType: 'json',
	data: {name: name, canvasSerial: canvasSerial},
	beforeSend: function(){
	},
	success: function(){
		getImages()
	}
	});
}

function loadImageToCanvas(name) {
	$.ajax({
	url: "/api?action=loadImageToCanvas",
	type: 'post',
	dataType: 'json',
	data: {name: name, canvasSerial: canvasSerial},
	beforeSend: function(){
	},
	success: function(json){
	}
	});
}

function loadAnimationFrameToCanvas(name, frame) {
	$.ajax({
	url: "/api?action=loadAnimationFrameToCanvas",
	type: 'post',
	dataType: 'json',
	data: {name: name, frame: frame},
	beforeSend: function(){
	},
	success: function(json){
	}
	});
}

function getImages() {
	$.ajax({
	url: "/api?action=getImages",
	type: 'post',
	dataType: 'html',
	data: {canvasSerial: canvasSerial},
	beforeSend: function(){
		$('#images').html('')
	},
	success: function(html){
		$('#images').html(html)
	}
	});
}

function clearDisplay(){
	$.ajax({
	url: "/api?action=clearDisplay&canvasSerial=" + canvasSerial,
	type: 'post',
	dataType: 'json',
	beforeSend: function(){
	},
	success: function(json){
		$.each(json, function(i,col){
			$.each(col, function(j, px){
				td = i +2;
				tr = j +2;
				$('#pixelTable tr:nth-child('+tr+') td:nth-child('+td+')').css('background-color','rgba('+px.R+','+px.G+','+px.B+',255)');
			});
		});
	}
	});
}

function toggleDrawer(e){
	target = $(e).data('target');
	$('#' + target).toggle();
	$(e).find('.fas').each(function(i,el){
			$(el).toggle();
		})
}

function pad(num, size) {
    var s = num+"";
    while (s.length < size) s = "0" + s;
    return s;
}

function refreshDisplayFromServer(){


	if(!drawmode){

		$.ajax({
		url: "/api?action=getDisplay&canvasSerial=" + canvasSerial,
		type: 'post',
		dataType: 'json',
		beforeSend: function(){
		},
		success: function(json){

			if(typeof json == "undefined"){
				return
			}

			canvasSerial = json.CanvasSerial;

			$.each(json.Canvas, function(i,col){
				$.each(col, function(j, px){

					ID = pad(i,2) + pad(j,2)

					//$('#pixelTable tr:nth-child('+tr+') td:nth-child('+td+')').css('background-color','rgba('+px.R+','+px.G+','+px.B+',255)');
					$('#px' + ID).css('background-color','rgba('+px.R+','+px.G+','+px.B+',255)');
				});
			});
		}
		});

	}
}

var thing;

function isNeighbor(pixelArray, target) {
	$.each(pixelArray, function(j,px){
			//console.log("Testing if " + target.X + "," + target.Y + " is neighbor of " + px.X + "," + px.Y);
			if((Math.abs(target.X - px.X) == 1 || target.X === px.X) && (Math.abs(target.Y - px.Y) == 1 || target.Y === px.Y)){
				return true;
			}
	});
	return false;
}

function setFill() {
	fill = true;
}

function setPixel(x,y){

	tr = y +2;
	td = x +2;
	target = $('#pixelTable tr:nth-child('+tr+') td:nth-child('+td+')');


	targetColor = $(target).css('background-color');

	//console.log("Target color: " + targetColor);

	thing = target;

	canvasMaxX = $('#pixelTable tr:nth-child(1) td').length -1;


	pixels = []
	px = {}
	px.X = x;
	px.Y = y;
	px.R = color.R;
	px.G = color.G;
	px.B = color.B;
	px.A = 255;
	pixels.push(px);


	if(fill === true){

	
		pxJSON = JSON.stringify(px);

		console.log(pxJSON);


		$.ajax({
		url: "/api?action=fillPixel",
		type: 'post',
		data: {px: pxJSON},
		dataType: 'json',
		beforeSend: function(){
		},
		success: function(json){

		}
		});

		fill = false;

		return
	}

	pxJSON = JSON.stringify(pixels);

	// Set the target to the specified color
	$.each(pixels, function(i, v){
		ltr = v.Y+2;
		ltd = v.X+2;
		$('#pixelTable tr:nth-child('+ltr+') td:nth-child('+ltd+')').css('background-color','rgba('+color.R+','+color.G+','+color.B+','+color.A+')');
	});
	




	$.ajax({
	url: "/api?action=setPixel",
	type: 'post',
	data: {px: pxJSON},
	dataType: 'json',
	beforeSend: function(){
	},
	success: function(json){

	}
	});
}