var isCollapsed = false;

function isCollapsedOnLoad() {
    if (window.localStorage.getItem('menuIsCollapsed') == 'yes') {
        return true;
    }
    return false;
}


function handleMenu(state) {
    if (state === true) {
        $('body').addClass('sidebar-collapse');
        isCollapsed = true;
    } else {
        $('body').removeClass('sidebar-collapse');
        isCollapsed = false;
    }
}

function handleMenuStorage(state) {
    if (state === true) {
        window.localStorage.removeItem('menuIsCollapsed');
    } else {
        window.localStorage.setItem('menuIsCollapsed', 'yes');
    }
}

function getMenuState(){
    return isCollapsed;
}

$(document).ready(function () {
    var state = isCollapsedOnLoad();
    handleMenu(state);

    $('.sidebar-toggle').click(function () {
        //console.log(getMenuState());
        handleMenuStorage(getMenuState());
        handleMenu(!getMenuState());
    })
});