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

namespace Statusengine\Loader\Mysql;

use Statusengine\Backend\StorageBackend;
use Statusengine\Loader\ServicePerfdataLoaderInterface;
use Statusengine\ValueObjects\ServicePerfdataQueryOptions;

class ServicePerfdataLoader implements ServicePerfdataLoaderInterface {

    /**
     * @var \Statusengine\Backend\Mysql\MySQL
     */
    private $Backend;

    /**
     * ServicePerfdataLoader constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend) {
        $this->Backend = $StorageBackend->getBackend();
    }


    /**
     * @param ServicePerfdataQueryOptions $ServicePerfdataQueryOptions
     * @return array
     */
    public function getServicePerfdata(ServicePerfdataQueryOptions $ServicePerfdataQueryOptions) {
        $query = $this->Backend->prepare('SELECT timestamp_unix as timestamp, value from statusengine_perfdata 
        WHERE hostname=? AND service_description=? AND label=? and timestamp_unix < ? and timestamp_unix > ? ORDER BY timestamp ASC');
        $query->bindValue(1, $ServicePerfdataQueryOptions->getHostname());
        $query->bindValue(2, $ServicePerfdataQueryOptions->getServicedescription());
        $query->bindValue(3, $ServicePerfdataQueryOptions->getMetric());
        $query->bindValue(4, $ServicePerfdataQueryOptions->getStart());
        $query->bindValue(5, $ServicePerfdataQueryOptions->getEnd());

        return $this->Backend->fetchAll($query);
    }

}
