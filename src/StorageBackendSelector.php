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

namespace Statusengine;


use Statusengine\Backend\Crate\Crate;
use Statusengine\Backend\Mysql\MySQL;
use Statusengine\Backend\StorageBackend;
use Statusengine\Exceptions\UnknownBackendException;

class StorageBackendSelector {

    /**
     * @var Config
     */
    private $Config;

    /**
     * StorageBackendSelector constructor.
     * @param Config $Config
     */
    public function __construct(Config $Config) {
        $this->Config = $Config;
    }

    /**
     * @return StorageBackend
     * @throws UnknownBackendException
     */
    public function getStorageBackend(){
        if($this->Config->isCrateEnabled()){
            return new StorageBackend(new Crate($this->Config), $this->Config);
        }

        if($this->Config->isMysqlEnabled()){
            return new StorageBackend(new MySQL($this->Config), $this->Config);
        }

        throw new UnknownBackendException('No storage backend enabled!');
    }

}
