<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2018  Daniel Ziegler
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

require __DIR__ . '/../../vendor/autoload.php';
define('DS', DIRECTORY_SEPARATOR);


use \Psr\Http\Message\ServerRequestInterface as Request;
use \Psr\Http\Message\ResponseInterface as Response;


// Some bootstrap stuff
$SessionHandler = new \Statusengine\SessionHandler();
$StatusengineConfig = new \Statusengine\Config();
$StorageBackendSelector = new \Statusengine\StorageBackendSelector($StatusengineConfig);

//Tell slim to print useful error messages
$config['displayErrorDetails'] = true;
$app = new \Slim\App(["settings" => $config]);

// Register with container
$container = $app->getContainer();
$container['SessionHandler'] = $SessionHandler;
$container['csrf'] = function ($c) {
    return new \Slim\Csrf\Guard;
};
$container['StorageBackend'] = $StorageBackendSelector->getStorageBackend();


$app->add(new \Statusengine\StatusengineAuth(
    $StatusengineConfig,
    $StorageBackendSelector->getStorageBackend()
));

// Register middleware for all routes
// If you are implementing per-route checks you must not add this
$app->add($container->get('csrf'));

//session_write_close();

/**
 * @var \Statusengine\Backend\StorageBackend $StorageBackend
 *
 * Paremters:
 * username string
 * password string
 * csrf_name
 * csrf_value
 *
 * Send a GET request first to get the values for csrf_name and csrf_value!
 * Send a post request and pass the values for username, password, csrf_name and csrf_value to login!
 */
$app->map(['GET', 'POST'], '/login', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');

    if ($request->isPost()) {
        return $response->withJson([
            'message' => 'Login Successfully'
        ]);
    }

    if ($request->isGet()) {

        $nameKey = $this->csrf->getTokenNameKey();
        $valueKey = $this->csrf->getTokenValueKey();
        $name = $request->getAttribute($nameKey);
        $value = $request->getAttribute($valueKey);

        $fields = [
            'username' => 'your_username',
            'password' => 'your_password',
            $nameKey   => $name,
            $valueKey  => $value
        ];

        return $response->withJson([
            'message'         => 'Send a POST request with the required fields to login',
            'required_fields' => $fields,
        ]);
    }

});

$app->get('/loginstate', function (Request $request, Response $response) {
    $Session = new \Statusengine\SessionHandler();
    $Config = new \Statusengine\Config();

    $isLoggedIn = false;
    if ($Session->has('loginSuccessfully')) {
        $isLoggedIn = $Session->get('loginSuccessfully');
    }

    return $response->withJson([
        'isLoggedIn'                => $isLoggedIn,
        'username'                  => $Session->getUsername(),
        'auth_type'                 => $Config->getAuthType(),
        'isAnonymousAllowed'        => $Config->isAnonymousAllowed(),
        'canAnonymousSubmitCommand' => $Config->canAnonymousSubmitCommand()
    ]);

});

$app->get('/logout', function (Request $request, Response $response) {
    $Session = new \Statusengine\SessionHandler();
    $Session->destroy();
    return $response->withJson([
        'message' => 'You have been logged out'
    ]);
});

/**
 * Paremters:
 * none
 */
$app->get('/', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();
    $DashboardQueryOptions = new \Statusengine\ValueObjects\DashboardQueryOptions($params);
    $DashboardController = new \Statusengine\Controller\Dashboard(
        $StorageBackend->getDashboardLoader()
    );
    $data = $DashboardController->index($DashboardQueryOptions);

    return $response->withJson($data);
});

/**
 * Paremters:
 * limit int
 * entry_time__gt > timestamp
 * entry_time__lt < timestamp
 * logentry_data__like string
 * cluster_name array
 */
$app->get('/logentries', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');

    $params = $request->getQueryParams();
    $LogentryController = new \Statusengine\Controller\Logentry(
        $StorageBackend->getLogentryLoader()
    );

    $data = $LogentryController->index(new \Statusengine\ValueObjects\LogentryQueryOptions($params));
    return $response->withJson($data);
});

/**
 * Parameters:
 * limit int
 * offset int
 * order [see HostQueryOptions::$columnsForOrder]
 * direction asc, desc
 * state array [up, down, unreachable]
 * hostname__like string
 */
$app->get('/hosts', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $HostController = new \Statusengine\Controller\Host(
        $StorageBackend->getHostLoader()
    );

    $data = $HostController->index(new \Statusengine\ValueObjects\HostQueryOptions($params));
    return $response->withJson($data);

});

/**
 * Parameters:
 * limit int
 * offset int
 * hostname__like string
 * servicedescription__like string
 * order: hostname,service_description string with separated by comma ,
 * state array [ok, warning, critical, unknown]
 */
$app->get('/services', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $ServiceController = new \Statusengine\Controller\Service(
        $StorageBackend->getServiceLoader()
    );

    $data = $ServiceController->index(new \Statusengine\ValueObjects\ServiceQueryOptions($params));
    $data = $ServiceController->keepOrder($data);
    return $response->withJson($data);

});

/**
 * Parameters:
 * limit int
 * offset int
 * hostname__like string
 * servicedescription__like string
 * order: hostname,service_description string with separated by comma (,)
 */
$app->get('/problems', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();
    $params['is_acknowledged'] = false;
    $params['is_in_downtime'] = false;

    $params['state'] = ['warning', 'critical', 'unknown'];

    $ServiceController = new \Statusengine\Controller\Service(
        $StorageBackend->getServiceLoader()
    );

    $data = $ServiceController->index(new \Statusengine\ValueObjects\ServiceQueryOptions($params));
    $data = $ServiceController->keepOrder($data);
    return $response->withJson($data);

});

/**
 * Parameters:
 * limit int
 * offset int
 */
$app->get('/globalproblems', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $ServiceController = new \Statusengine\Controller\Service(
        $StorageBackend->getServiceLoader()
    );

    $data = $ServiceController->problems(new \Statusengine\ValueObjects\ServiceQueryOptions($params));
    return $response->withJson($data);

});

$app->get('/cluster', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');

    $ClusterController = new \Statusengine\Controller\Cluster(
        $StorageBackend->getClusterLoader()
    );

    $data = $ClusterController->index();
    return $response->withJson($data);

});

$app->get('/clusteroverview', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');

    $ClusterController = new \Statusengine\Controller\Cluster(
        $StorageBackend->getClusterLoader()
    );

    $data = $ClusterController->overview();
    return $response->withJson($data);

});

/**
 * Parameters:
 * hostname string
 * servicedescription__like string
 * service_state array [ok, warning, critical, unknown]
 */
$app->get('/hostdetails', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $HostController = new \Statusengine\Controller\Host(
        $StorageBackend->getHostLoader()
    );

    $data = $HostController->hostdetails(new \Statusengine\ValueObjects\HostQueryOptions($params));

    $Config = new \Statusengine\Config();
    $data['external_urls'] = $Config->getExternalUrls();

    return $response->withJson($data);

});

/**
 * Parameters:
 * hostname string
 * servicedescription string
 */
$app->get('/servicedetails', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $ServiceController = new \Statusengine\Controller\Service(
        $StorageBackend->getServiceLoader()
    );

    $Config = new \Statusengine\Config();

    $data = $ServiceController->servicedetails(new \Statusengine\ValueObjects\ServiceQueryOptions($params));
    $data['display_perfdata'] = $Config->getDisplayPerfdata();
    $data['external_urls'] = $Config->getExternalUrls();
    return $response->withJson($data);

});

/**
 * Parameters:
 * hostname string
 * servicedescription string
 * metric string
 * points int
 * start  int (timestamp)
 * end    int (timestamp)
 * compression_algorithm string (avg, min, max)
 */
$app->get('/serviceperfdata', function (Request $request, Response $response) {
    $params = $request->getQueryParams();


    $Config = new \Statusengine\Config();
    $StorageBackend = new \Statusengine\Backend\StorageBackend(new \Statusengine\Backend\Crate\Crate($Config), $Config);

    $ServicePerfdataController = new \Statusengine\Controller\ServicePerfdata(
        $StorageBackend->getServicePerfdataLoader()
    );

    $data = $ServicePerfdataController->index(new \Statusengine\ValueObjects\ServicePerfdataQueryOptions($params));
    return $response->withJson($data);

});

/**
 * Parameters:
 * hostname string
 * servicedescription string
 * command_id int
 */
$app->get('/externalcommand', function (Request $request, Response $response) {
    $Config = new \Statusengine\Config();
    $Session = new \Statusengine\SessionHandler();
    $isLoggedIn = false;
    if ($Session->has('loginSuccessfully')) {
        $isLoggedIn = $Session->get('loginSuccessfully');
    }
    if ($isLoggedIn === false && $Config->canAnonymousSubmitCommand() === false) {
        return $response->withJson(['message' => 'External commands are disabled for anonymous users']);
    }


    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $ExternalCommandController = new \Statusengine\Controller\ExternalCommand(
        $StorageBackend->getExternalCommandSaver()
    );

    $result = $ExternalCommandController->index(new \Statusengine\ValueObjects\ExternalCommandQueryOptions($params));
    return $response->withJson(['result' => $result]);

});

/**
 * Parameters:
 * hostname string
 * servicedescription string
 * command_name string
 * command_args_n (required command arguments for your command)
 */
$app->get('/externalcommand_args', function (Request $request, Response $response) {
    $Config = new \Statusengine\Config();
    $Session = new \Statusengine\SessionHandler();
    $isLoggedIn = false;
    if ($Session->has('loginSuccessfully')) {
        $isLoggedIn = $Session->get('loginSuccessfully');
    }
    if ($isLoggedIn === false && $Config->canAnonymousSubmitCommand() === false) {
        return $response->withJson(['message' => 'External commands are disabled for anonymous users']);
    }

    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $SessionHandler = new \Statusengine\SessionHandler();
    $authorName = 'Anonymous';
    if ($SessionHandler->has('username')) {
        $authorName = $SessionHandler->get('username');
    }

    $params['author_name'] = $authorName;

    $ExternalCommandController = new \Statusengine\Controller\ExternalCommand(
        $StorageBackend->getExternalCommandSaver()
    );

    $result = $ExternalCommandController->args(new \Statusengine\ValueObjects\ExternalCommandArgsQueryOptions($params));
    return $response->withJson(['result' => $result]);

});

/**
 * Parameters:
 * limit int
 * offset int
 * order [see HostCheckQueryOptions::$columnsForOrder]
 * direction asc, desc
 * state array [up, down, unreachable]
 * output__like string
 * hostname string
 */
$app->get('/hostchecks', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $HostCheckController = new \Statusengine\Controller\HostCheck(
        $StorageBackend->getHostCheckLoader()
    );

    $data = $HostCheckController->index(new \Statusengine\ValueObjects\HostCheckQueryOptions($params));
    return $response->withJson($data);

});

/**
 * Parameters:
 * limit int
 * offset int
 * order [see HostStateHistoryQueryOptions::$columnsForOrder]
 * direction asc, desc
 * state array [up, down, unreachable]
 * output__like string
 * hostname string
 */
$app->get('/hoststatehistory', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $HostStateHistoryController = new \Statusengine\Controller\HostStateHistory(
        $StorageBackend->getHostStateHistoryLoader()
    );

    $data = $HostStateHistoryController->index(new \Statusengine\ValueObjects\HostStateHistoryQueryOptions($params));
    return $response->withJson($data);

});

/**
 * Parameters:
 * hostname string
 * limit int
 * offset int
 * order [see HostNotificationsQueryOptions::$columnsForOrder]
 * direction asc, desc
 * state array [up, down, unreachable]
 * output__like string
 */
$app->get('/hostnotifications', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $HostNotificationController = new \Statusengine\Controller\HostNotification(
        $StorageBackend->getHostNotificationLoader()
    );


    $data = $HostNotificationController->index(new \Statusengine\ValueObjects\HostNotificationQueryOptions($params));
    return $response->withJson($data);

});

/**
 * Parameters:
 * hostname string
 * limit int
 * offset int
 * order [see HostAcknowledgementQueryOptions::$columnsForOrder]
 * direction asc, desc
 * state array [up, down, unreachable]
 * comment_data__like string
 * entry_time__lt < timestamp
 * entry_time__gt > timestamp
 */
$app->get('/hostacknowledgements', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $AcknowledgementController = new \Statusengine\Controller\HostAcknowledgement(
        $StorageBackend->getHostAcknowledgementLoader()
    );

    $data = $AcknowledgementController->index(new \Statusengine\ValueObjects\HostAcknowledgementQueryOptions($params));
    return $response->withJson($data);

});

/**
 * Parameters:
 * limit int
 * offset int
 * order [see ServiceCheckQueryOptions::$columnsForOrder]
 * direction asc, desc
 * state array [ok, warning, critical, unknown]
 * output__like string
 * hostname string
 * service_description string
 */
$app->get('/servicechecks', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $ServiceCheckController = new \Statusengine\Controller\ServiceCheck(
        $StorageBackend->getServiceCheckLoader()
    );

    $data = $ServiceCheckController->index(new \Statusengine\ValueObjects\ServiceCheckQueryOptions($params));
    return $response->withJson($data);

});

/**
 * Parameters:
 * limit int
 * offset int
 * order [see ServiceStateHistoryQueryOptions::$columnsForOrder]
 * direction asc, desc
 * state array [ok, warning, critical, unknown]
 * output__like string
 * hostname string
 * service_description string
 */
$app->get('/servicestatehistory', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $ServiceStateHistoryController = new \Statusengine\Controller\ServiceStateHistory(
        $StorageBackend->getServiceStateHistoryLoader()
    );

    $data = $ServiceStateHistoryController->index(new \Statusengine\ValueObjects\ServiceStateHistoryQueryOptions($params));
    return $response->withJson($data);

});

/**
 * Parameters:
 * hostname string
 * servicedescription string
 * limit int
 * offset int
 * order [see HostAcknowledgementQueryOptions::$columnsForOrder]
 * direction asc, desc
 * state array [up, down, unreachable]
 * comment_data__like string
 * entry_time__lt < timestamp
 * entry_time__gt > timestamp
 */
$app->get('/serviceacknowledgements', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $AcknowledgementController = new \Statusengine\Controller\ServiceAcknowledgement(
        $StorageBackend->getServiceAcknowledgementLoader()
    );

    $data = $AcknowledgementController->index(new \Statusengine\ValueObjects\ServiceAcknowledgementQueryOptions($params));
    return $response->withJson($data);

});

/**
 * Parameters:
 * hostname string
 * servicedescription string
 * limit int
 * offset int
 * order [see HostNotificationsQueryOptions::$columnsForOrder]
 * direction asc, desc
 * state array [up, down, unreachable]
 * output__like string
 */
$app->get('/servicenotifications', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $HostNotificationController = new \Statusengine\Controller\ServiceNotification(
        $StorageBackend->getServiceNotificationLoader()
    );


    $data = $HostNotificationController->index(new \Statusengine\ValueObjects\ServiceNotificationQueryOptions($params));
    return $response->withJson($data);

});

/**
 * Parameters:
 * hostname__like string
 * limit int
 */
$app->get('/hostsearch', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $HostController = new \Statusengine\Controller\Host(
        $StorageBackend->getHostLoader()
    );

    $data = $HostController->search(new \Statusengine\ValueObjects\HostSearchQueryOptions($params));
    return $response->withJson($data);

});

/**
 * Parameters:
 * hostname string
 * servicedescription string
 */
$app->get('/servicedowntime', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $ServiceDowntime = new \Statusengine\Controller\ServiceDowntime(
        $StorageBackend->getServiceDowntimeLoader()
    );


    $data = $ServiceDowntime->index(new \Statusengine\ValueObjects\ServiceDowntimeQueryOptions($params));
    return $response->withJson($data);

});

/**
 * Parameters:
 * hostname string
 */
$app->get('/hostdowntime', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $HostDowntime = new \Statusengine\Controller\HostDowntime(
        $StorageBackend->getHostdowntimeLoader()
    );


    $data = $HostDowntime->index(new \Statusengine\ValueObjects\HostDowntimeQueryOptions($params));
    return $response->withJson($data);

});

/**
 * Parameters:
 * hostname__like string
 * servicedescription__like string
 * object_type string (host/service)
 */
$app->get('/scheduleddowntimes', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $ScheduleddowntimeQueryOptions = new \Statusengine\ValueObjects\ScheduleddowntimeQueryOptions($params);

    if ($ScheduleddowntimeQueryOptions->isHostRequest() === true) {
        $ScheduleddowntimeHost = new \Statusengine\Controller\ScheduleddowntimeHost(
            $StorageBackend->getScheduleddowntimeHostLoader()
        );

        $data = $ScheduleddowntimeHost->index($ScheduleddowntimeQueryOptions);
        return $response->withJson($data);
    }

    if ($ScheduleddowntimeQueryOptions->isHostRequest() === false) {
        $ScheduleddowntimeService = new \Statusengine\Controller\ScheduleddowntimeService(
            $StorageBackend->getScheduleddowntimeServiceLoader()
        );

        $data = $ScheduleddowntimeService->index($ScheduleddowntimeQueryOptions);
        return $response->withJson($data);
    }

});

/**
 * Parameters:
 * object_type string (host/service)
 * hostname__like string
 * servicedescription__like string
 * limit int
 * offset int
 * entry_time__lt < timestamp
 * entry_time__gt > timestamp
 */
$app->get('/acknowledgements', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();

    $QueryOptions = new \Statusengine\ValueObjects\AcknowledgementQueryOptions($params);

    if ($QueryOptions->isHostRequest()) {
        $AcknowledgementController = new \Statusengine\Controller\HostAcknowledgement(
            $StorageBackend->getHostAcknowledgementLoader()
        );

        $data = $AcknowledgementController->getCurrentAcknowledgements(
            new \Statusengine\ValueObjects\HostAcknowledgementQueryOptions($params)
        );
        return $response->withJson($data);

    }

    $AcknowledgementController = new \Statusengine\Controller\ServiceAcknowledgement(
        $StorageBackend->getServiceAcknowledgementLoader()
    );

    $data = $AcknowledgementController->getCurrentAcknowledgements(
        new \Statusengine\ValueObjects\ServiceAcknowledgementQueryOptions($params)
    );
    return $response->withJson($data);
});

/**
 * will always return 0 for hosts up and services ok!
 * Parameters:
 * hide_ack_and_downtime (string true/false)
 */
$app->get('/menustats', function (Request $request, Response $response) {
    $StorageBackend = $this->get('StorageBackend');
    $params = $request->getQueryParams();
    $DashboardQueryOptions = new \Statusengine\ValueObjects\DashboardQueryOptions($params);
    $DashboardQueryOptions->setHostStates([1, 2]); //DO NOT PASS ARGUMENTS FROM $_GET OR $_POST!!!
    $DashboardQueryOptions->setServiceStates([1, 2, 3]);//DO NOT PASS ARGUMENTS FROM $_GET OR $_POST!!!
    $DashboardController = new \Statusengine\Controller\Dashboard(
        $StorageBackend->getDashboardLoader()
    );
    $data = $DashboardController->menuStats($DashboardQueryOptions);

    return $response->withJson($data);
});

$app->run();
