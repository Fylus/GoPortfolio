const burgermenu = $('#burgermenu');
const hamburgercontainer = $('#hamburger-container')
const hamburger = $('#hamburger');
const links = $("#burgermenu a");

let menuShown = false;

hamburgercontainer.click(function () {
    if (menuShown) {
        burgermenu.addClass('hidden');
        hamburger.removeClass("is-active")
        menuShown = false;
    } else {
        burgermenu.removeClass('hidden');
        hamburger.addClass("is-active")
        menuShown = true;
    }
});

links.click(function () {
        if (menuShown) {
            hamburgercontainer.click();
        }
});
