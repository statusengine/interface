<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2017  Daniel Ziegler
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

namespace Statusengine\Backend;

use Statusengine\Backend\Crate\Crate;
use Statusengine\Backend\Mysql\MySQL;
use Statusengine\Config;
use Statusengine\Exceptions\UnknownBackendException;
use Statusengine\Loader\Crate\ClusterLoader;
use Statusengine\Loader\Crate\DashboardLoader;
use Statusengine\Loader\Crate\ExternalCommandSaver;
use Statusengine\Loader\Crate\HostAcknowledgementLoader;
use Statusengine\Loader\Crate\HostCheckLoader;
use Statusengine\Loader\Crate\HostDowntimeLoader;
use Statusengine\Loader\Crate\HostLoader;
use Statusengine\Loader\Crate\HostNotificationLoader;
use Statusengine\Loader\Crate\HostStateHistoryLoader;
use Statusengine\Loader\Crate\LogentryLoader;
use Statusengine\Loader\Crate\ScheduleddowntimeHostLoader;
use Statusengine\Loader\Crate\ScheduleddowntimeServiceLoader;
use Statusengine\Loader\Crate\ServiceAcknowledgementLoader;
use Statusengine\Loader\Crate\ServiceCheckLoader;
use Statusengine\Loader\Crate\ServiceDowntimeLoader;
use Statusengine\Loader\Crate\ServiceLoader;
use Statusengine\Loader\Crate\ServiceNotificationLoader;
use Statusengine\Loader\Crate\ServicePerfdataLoader;
use Statusengine\Loader\Crate\ServiceStateHistoryLoader;
use Statusengine\Loader\Crate\UserLoader;

class StorageBackend {

    private $Backend;

    /**
     * @var Config
     */
    private $Config;

    /**
     * @var string
     */
    private $backendClassName;

    /**
     * StorageBackend constructor.
     * @param $Backend
     */
    public function __construct($Backend, Config $Config) {
        $this->Backend = $Backend;
        $this->Config = $Config;
        $this->backendClassName = get_class($this->Backend);
    }

    /**
     * @return Crate
     */
    public function getBackend() {
        return $this->Backend;
    }

    /**
     * @return DashboardLoader|\Statusengine\Loader\Mysql\DashboardLoader
     * @throws UnknownBackendException
     */
    public function getDashboardLoader() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new DashboardLoader($this);
        }

        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\DashboardLoader($this);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

    /**
     * @return LogentryLoader|\Statusengine\Loader\Mysql\LogentryLoader
     * @throws UnknownBackendException
     */
    public function getLogentryLoader() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new LogentryLoader($this);
        }

        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\LogentryLoader($this);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

    /**
     * @return HostLoader|\Statusengine\Loader\Mysql\HostLoader
     * @throws UnknownBackendException
     */
    public function getHostLoader() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new HostLoader($this);
        }
        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\HostLoader($this);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

    /**
     * @return ServiceLoader|\Statusengine\Loader\Mysql\ServiceLoader
     * @throws UnknownBackendException
     */
    public function getServiceLoader() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new ServiceLoader($this);
        }
        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\ServiceLoader($this);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

    /**
     * @return ClusterLoader|\Statusengine\Loader\Mysql\ClusterLoader
     * @throws UnknownBackendException
     */
    public function getClusterLoader() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new ClusterLoader($this);
        }
        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\ClusterLoader($this);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

    /**
     * @return ServicePerfdataLoader|\Statusengine\Loader\Elasticsearch\ServicePerfdataLoader|\Statusengine\Loader\Graphite\ServicePerfdataLoader|\Statusengine\Loader\Mysql\ServicePerfdataLoader
     * @throws UnknownBackendException
     */
    public function getServicePerfdataLoader() {
        if ($this->Config->getPerfdataBackend() === 'graphite') {
            return new \Statusengine\Loader\Graphite\ServicePerfdataLoader($this->Config);
        }

        if ($this->Config->getPerfdataBackend() === 'crate') {
            $StorageBackend = new StorageBackend(new Crate($this->Config), $this->Config);
            return new ServicePerfdataLoader($StorageBackend);
        }

        if ($this->Config->getPerfdataBackend() === 'mysql') {
            $StorageBackend = new StorageBackend(new MySQL($this->Config), $this->Config);
            return new \Statusengine\Loader\Mysql\ServicePerfdataLoader($StorageBackend);
        }

        if ($this->Config->getPerfdataBackend() === 'elasticsearch') {
            return new \Statusengine\Loader\Elasticsearch\ServicePerfdataLoader($this->Config);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));

    }

    /**
     * @return ExternalCommandSaver|\Statusengine\Loader\Mysql\ExternalCommandSaver
     * @throws UnknownBackendException
     */
    public function getExternalCommandSaver() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new ExternalCommandSaver($this);
        }
        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\ExternalCommandSaver($this);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

    /**
     * @return HostCheckLoader|\Statusengine\Loader\Mysql\HostCheckLoader
     * @throws UnknownBackendException
     */
    public function getHostCheckLoader() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new HostCheckLoader($this);
        }
        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\HostCheckLoader($this);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

    /**
     * @return HostStateHistoryLoader|\Statusengine\Loader\Mysql\HostStateHistoryLoader
     * @throws UnknownBackendException
     */
    public function getHostStateHistoryLoader() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new HostStateHistoryLoader($this);
        }

        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\HostStateHistoryLoader($this);
        }


        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

    /**
     * @return HostNotificationLoader|\Statusengine\Loader\Mysql\HostNotificationLoader
     * @throws UnknownBackendException
     */
    public function getHostNotificationLoader() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new HostNotificationLoader($this);
        }

        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\HostNotificationLoader($this);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

    /**
     * @return HostAcknowledgementLoader|\Statusengine\Loader\Mysql\HostAcknowledgementLoader
     * @throws UnknownBackendException
     */
    public function getHostAcknowledgementLoader() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new HostAcknowledgementLoader($this);
        }

        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\HostAcknowledgementLoader($this);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

    /**
     * @return ServiceCheckLoader|\Statusengine\Loader\Mysql\ServiceCheckLoader
     * @throws UnknownBackendException
     */
    public function getServiceCheckLoader() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new ServiceCheckLoader($this);
        }

        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\ServiceCheckLoader($this);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

    /**
     * @return ServiceStateHistoryLoader|\Statusengine\Loader\Mysql\ServiceStateHistoryLoader
     * @throws UnknownBackendException
     */
    public function getServiceStateHistoryLoader() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new ServiceStateHistoryLoader($this);
        }

        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\ServiceStateHistoryLoader($this);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

    /**
     * @return ServiceAcknowledgementLoader|\Statusengine\Loader\Mysql\ServiceAcknowledgementLoader
     * @throws UnknownBackendException
     */
    public function getServiceAcknowledgementLoader() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new ServiceAcknowledgementLoader($this);
        }

        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\ServiceAcknowledgementLoader($this);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

    /**
     * @return ServiceNotificationLoader|\Statusengine\Loader\Mysql\ServiceNotificationLoader
     * @throws UnknownBackendException
     */
    public function getServiceNotificationLoader() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new ServiceNotificationLoader($this);
        }

        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\ServiceNotificationLoader($this);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

    /**
     * @return UserLoader|\Statusengine\Loader\Mysql\UserLoader
     * @throws UnknownBackendException
     */
    public function getUserLoader() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new UserLoader($this);
        }

        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\UserLoader($this);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

    /**
     * @return ServiceDowntimeLoader|\Statusengine\Loader\Mysql\ServiceDowntimeLoader
     * @throws UnknownBackendException
     */
    public function getServiceDowntimeLoader() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new ServiceDowntimeLoader($this);
        }

        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\ServiceDowntimeLoader($this);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

    /**
     * @return HostDowntimeLoader|\Statusengine\Loader\Mysql\HostDowntimeLoader
     * @throws UnknownBackendException
     */
    public function getHostdowntimeLoader() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new HostDowntimeLoader($this);
        }

        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\HostDowntimeLoader($this);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

    /**
     * @return ScheduleddowntimeHostLoader|\Statusengine\Loader\Mysql\ScheduleddowntimeHostLoader
     * @throws UnknownBackendException
     */
    public function getScheduleddowntimeHostLoader() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new ScheduleddowntimeHostLoader($this);
        }

        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\ScheduleddowntimeHostLoader($this);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

    /**
     * @return ScheduleddowntimeServiceLoader|\Statusengine\Loader\Mysql\ScheduleddowntimeServiceLoader
     * @throws UnknownBackendException
     */
    public function getScheduleddowntimeServiceLoader() {
        if ($this->backendClassName == 'Statusengine\Backend\Crate\Crate') {
            return new ScheduleddowntimeServiceLoader($this);
        }

        if ($this->backendClassName == 'Statusengine\Backend\Mysql\MySQL') {
            return new \Statusengine\Loader\Mysql\ScheduleddowntimeServiceLoader($this);
        }

        throw new UnknownBackendException(sprintf('Storage Backend \'%s\' is unknown', $this->backendClassName));
    }

}
